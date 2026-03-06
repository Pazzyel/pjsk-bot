
from __future__ import annotations

from typing import Any, Dict

import asyncio
import os,sys

from pathlib import Path

from langchain_classic.agents import AgentExecutor, create_tool_calling_agent
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_mcp_adapters.client import MultiServerMCPClient
from langchain_openai import ChatOpenAI

from .pjsk_tools import get_chart, get_jacket


class PJSKAgent:
    def __init__(self) -> None:
        tools_server_path = str(Path(__file__).with_name("pjsk_tools.py").resolve())
        self.mcp_client = MultiServerMCPClient(
            {
                "pjsk": {
                    "transport": "stdio",
                    "command": sys.executable,
                    "args": [tools_server_path],
                }
            }
        )
        self.tools = asyncio.run(self.mcp_client.get_tools(server_name="pjsk"))
        self.model = ChatOpenAI(
            model="GLM-5",
            temperature=0.2,
            base_url="https://api.edgefn.net/v1",
            api_key=os.environ["OPENAI_API_KEY"]
        )
        self.system_prompt = (
            "你是一个PJSK（游戏：世界计划：缤纷舞台）助手。"
            "如果用户问题和PJSK无关，则直接正常回答。也不需要特地提到PJSK"
            "如果用户需要歌曲信息，必须调用search_music工具。"
            "当拿到歌曲列表后，只返回第一首，从返回的JSON解析出歌曲的信息提供给用户，输出的格式必须严格为：\n"
            "歌曲id: <value>\n"
            "原始名称: <value>\n"
            "中文名称: <value>\n"
            "作词: <value>\n"
            "做曲: <value>\n"
            "谱曲: <value>\n"
            "难度:\n"
            "  easy: <value>\n"
            "  normal: <value>\n"
            "  hard: <value>\n"
            "  expert: <value>\n"
            "  master: <value>\n"
            "  append: <value>\n"
            "如果是谱面或曲绘请求，优先调用对应工具。"
            "如果用户需要歌曲的曲绘，直接返回：Jacket({id})，其中{id}你需要替换成从用户的提问中抽取的歌曲id。"
            "如果用户需要歌曲的谱面，直接返回：Chart({id},{level})，其中{id}你需要替换成从用户的提问中抽取的歌曲id，"
            "{level}你需要替换成从用户的提问中理解出的歌曲难度，只能是easy,normal,hard,expert,append之一"
        )
        prompt = ChatPromptTemplate.from_messages(
            [
                ("system", "{system_prompt}"),
                ("human", "{input}"),
                MessagesPlaceholder("agent_scratchpad"),
            ]
        )
        agent = create_tool_calling_agent(self.model, self.tools, prompt)
        self.agent_executor = AgentExecutor(
            agent=agent,
            tools=self.tools,
            max_iterations=5,
            handle_parsing_errors=True,
            verbose=False,
        )

    async def ask(self, user_input: str) -> str | bytes:
        text = user_input.strip()
        print("Agent收到了问题：" + user_input)
        if not text:
            return "Please provide a question."

        try:
            result = await self.agent_executor.ainvoke(
                {
                    "input": text,
                    "system_prompt": self.system_prompt,
                }
            )
            content = result.get("output", "")
            content = content if isinstance(content, str) else str(content)
            if (content.startswith("Jacket(") and content.endswith(")")):
                song_id: str = content[len("Jacket("):len(content) - 1]
                return get_jacket(song_id=song_id)
            elif (content.startswith("Chart(") and content.endswith(")")):
                temp = content[len("Chart("):len(content) - 1]
                params: list[str] = temp.split(",")
                return get_chart(song_id=params[0],level=params[1])
            else:
                return content
        except Exception as e:
            return repr(e)



    # This function was unused
    def _format_music_info(self, music: Dict[str, Any]) -> str:
        difficulties = music.get("difficulties") or {}
        lines = [
            f"歌曲id: {music.get('id', '')}",
            f"原始名称: {music.get('title', '')}",
            f"中文名称: {music.get('chinese_title', '')}",
            f"作词: {music.get('lyricist', '')}",
            f"做曲: {music.get('composer', '')}",
            f"谱曲: {music.get('arranger', '')}",
            "难度:",
        ]
        for key in ["easy", "normal", "hard", "expert", "master", "append"]:
            item = difficulties.get(key) or {}
            lines.append(f"  {key}: {item.get('playLevel', '')}")
        return "\n".join(lines)


if __name__ == "__main__":
    agent = PJSKAgent()
    while True:
        try:
            text = input("You> ").strip()
        except (EOFError, KeyboardInterrupt):
            print("\nBye")
            break
        if text.lower() in {"exit", "quit"}:
            print("Bye")
            break
        result = asyncio.run(agent.ask(text))
        if isinstance(result, bytes):
            print(f"Agent> [bytes] len={len(result)}")
        else:
            print(f"Agent> {result}")
