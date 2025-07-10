from fastapi import FastAPI, File, Form, UploadFile, HTTPException
from pydantic import BaseModel
import whisper
import tempfile
import os

app = FastAPI(title="Whisper ASR Service")
model = whisper.load_model(os.getenv("WHISPER_MODEL", "turbo"))

class TranscribeResponse(BaseModel):
    text: str
    duration_sec: float
    language: str

@app.post("/", response_model=TranscribeResponse)

async def transcribe(id: str = Form(...), file: UploadFile = File(...)):
    if not file.content_type.startswith("audio/"):
        raise HTTPException(status_code=400, detail="Not an audio file")

    suffix = os.path.splitext(file.filename)[1] or ".wav"
    with tempfile.NamedTemporaryFile(suffix=suffix, delete=True) as tmp:
        contents = await file.read()
        tmp.write(contents)
        tmp.flush()

        result = model.transcribe(tmp.name)

    return TranscribeResponse(
        text=result["text"],
        duration_sec=result["duration"],
        language=result["language"],
    )


