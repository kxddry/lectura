import whisper
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import requests

app = FastAPI()

model = whisper.load_model("turbo")

class TranscribeRequest(BaseModel):
    s3_url: str

@app.post("/asr")
def transcribe(req: TranscribeRequest):
    audio_bytes = requests.get(req.s3_url).content