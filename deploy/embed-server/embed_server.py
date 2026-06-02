import os

from fastapi import FastAPI
from pydantic import BaseModel, Field
from sentence_transformers import SentenceTransformer


MODEL_NAME = os.getenv("EMBED_MODEL", "BAAI/bge-large-zh-v1.5")
DEVICE = os.getenv("EMBED_DEVICE", "cpu")
BATCH_SIZE = int(os.getenv("EMBED_BATCH_SIZE", "32"))

app = FastAPI(title="Campus QA Embedding Server")
model = SentenceTransformer(MODEL_NAME, device=DEVICE)


class EmbedReq(BaseModel):
    texts: list[str] = Field(..., min_length=1)


@app.get("/healthz")
async def healthz():
    return {
        "status": "ok",
        "model": MODEL_NAME,
        "device": DEVICE,
    }


@app.post("/embed")
async def embed(req: EmbedReq):
    embeddings = model.encode(
        req.texts,
        normalize_embeddings=True,
        batch_size=BATCH_SIZE,
    )
    return {
        "model": MODEL_NAME,
        "dimensions": len(embeddings[0]) if len(embeddings) > 0 else 0,
        "embeddings": embeddings.tolist(),
    }
