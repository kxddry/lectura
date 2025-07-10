from fastapi import FastAPI, File, UploadFile
from pydantic import BaseModel
import whisper
import tempfile
import os

app = FastAPI(title="Whisper ASR Service")
model = whisper.load_model(os.getenv("WHISPER_MODEL", "turbo"))

class TranscribeResponse(BaseModel):
    text: str
    language: str

@app.post("/", response_model=TranscribeResponse)

def transcribe(file: UploadFile = File(...)):
    suffix = os.path.splitext(file.filename)[1] or ".wav"
    print(f"Transcribing {file.filename} to {suffix}")
    with tempfile.NamedTemporaryFile(suffix=suffix, delete=True) as tmp:
        print(f"Writing to temporary file {tmp.name}")
        contents = file.file.read()
        tmp.write(contents)
        tmp.flush()
        print(f"model transcribes {tmp.name}")
        result = model.transcribe(tmp.name)
        print(f"Transcription result: {result}")

    return TranscribeResponse(
        text=result["text"],
        language=result["language"],
    )


