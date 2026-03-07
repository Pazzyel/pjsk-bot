from __future__ import annotations

import os
from functools import lru_cache
from pathlib import Path
from typing import Any

import yaml


DEFAULT_SERVER_BASE_URL = "http://127.0.0.1:9470"
DEFAULT_LLM_MODEL = "GLM-5"
DEFAULT_LLM_BASE_URL = "https://api.edgefn.net/v1"


def _config_path() -> Path:
    env_path = os.getenv("PJSK_CONFIG_FILE", "").strip()
    if env_path:
        return Path(env_path).expanduser().resolve()
    return Path(__file__).with_name("application-docker.yaml")


def _read_yaml(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}
    with path.open("r", encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    return data if isinstance(data, dict) else {}


@lru_cache(maxsize=1)
def get_server_base_url() -> str:
    config = _read_yaml(_config_path())
    server = config.get("server", {})
    if isinstance(server, dict):
        base_url = str(server.get("base_url", "")).strip()
        if base_url:
            return base_url.rstrip("/")
    return DEFAULT_SERVER_BASE_URL


@lru_cache(maxsize=1)
def get_llm_model() -> str:
    config = _read_yaml(_config_path())
    llm = config.get("llm", {})
    if isinstance(llm, dict):
        model = str(llm.get("model", "")).strip()
        if model:
            return model
    return DEFAULT_LLM_MODEL


@lru_cache(maxsize=1)
def get_llm_base_url() -> str:
    config = _read_yaml(_config_path())
    llm = config.get("llm", {})
    if isinstance(llm, dict):
        base_url = str(llm.get("base_url", "")).strip()
        if base_url:
            return base_url.rstrip("/")
    return DEFAULT_LLM_BASE_URL
