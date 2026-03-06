from __future__ import annotations

import os
from typing import Any, Dict, Optional

import requests
from fastmcp import FastMCP


mcp = FastMCP("pjsk-tools")


def _server_base_url() -> str:
    return os.getenv("PJSK_SERVER_BASE_URL", "http://127.0.0.1:9470").rstrip("/")


@mcp.tool
def search_music(
    title: str = "",
    author: str = "",
    min_level: Optional[int] = None,
    max_level: Optional[int] = None,
    page_no: Optional[int] = None,
    page_size: Optional[int] = None,
) -> Dict[str, Any]:
    """Search songs from PJSK server API.

    Calls:
        GET /pjsk/search

    Args:
        title: Match keyword for `title` and `chinese_title`.
        author: Match keyword for `lyricist`, `composer`, `arranger`.
        min_level: Optional minimum difficulty.
        max_level: Optional maximum difficulty.
        page_no: Optional page number.
        page_size: Optional page size.

    Returns:
        JSON dict from server: usually {"total": int, "data": [music_info...]}.
    """

    params: Dict[str, Any] = {}
    if title:
        params["title"] = title
    if author:
        params["author"] = author
    if min_level is not None:
        params["min_level"] = min_level
    if max_level is not None:
        params["max_level"] = max_level
    if page_no is not None:
        params["page_no"] = page_no
    if page_size is not None:
        params["page_size"] = page_size

    resp = requests.get(f"{_server_base_url()}/pjsk/search", params=params, timeout=30)
    resp.raise_for_status()
    return resp.json()


def get_chart(song_id: str, level: str) -> bytes:
    """Fetch chart image bytes.

    Calls:
        GET /pjsk/charts?id=<song_id>&level=<level>
    """

    resp = requests.get(
        f"{_server_base_url()}/pjsk/charts",
        params={"id": song_id, "level": level},
        timeout=30,
    )
    resp.raise_for_status()
    return resp.content


def get_jacket(song_id: str) -> bytes:
    """Fetch jacket image bytes.

    Calls:
        GET /pjsk/jackets?id=<song_id>
    """

    resp = requests.get(
        f"{_server_base_url()}/pjsk/jackets",
        params={"id": song_id},
        timeout=30,
    )
    resp.raise_for_status()
    return resp.content


@mcp.tool
def update_music() -> Dict[str, Any]:
    """Trigger full music refresh into Elasticsearch.

    Calls:
        POST /pjsk/update
    """

    resp = requests.post(f"{_server_base_url()}/pjsk/update", timeout=120)
    resp.raise_for_status()
    return resp.json()
