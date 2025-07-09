from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import whisper
import requests
import tempfile
import os

app = FastAPI(title="Whisper ASR Service")

model = whisper.load_model(os.getenv("WHISPER_MODEL", "turbo"))

class TranscribeRequest(BaseModel):
    id: str
    audio_url: str

class TranscribeResponse(BaseModel):
    id: str
    text: str
    duration_sec: float
    language: str

@app.post("/transcribe", response_model=TranscribeResponse)
def transcribe(req: TranscribeRequest):
    r = requests.get(req.audio_url)
    if r.status_code != 200:
        raise HTTPException(status_code=403, detail=f"Failed to download audio: {r.status_code}")
    with tempfile.NamedTemporaryFile(suffix=".wav") as f:
        f.write(r.content)
        f.flush()
        result = model.transcribe(f.name)
    return TranscribeResponse(
        id=req.id,
        text=result["text"],
        duration_sec=result["duration"],
        language=result["language"],
    )
