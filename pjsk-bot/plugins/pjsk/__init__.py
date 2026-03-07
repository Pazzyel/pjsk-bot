import base64
from nonebot import get_driver, on_notice, on_message
from nonebot.plugin import PluginMetadata, on_command
import requests, psutil
from nonebot.adapters.onebot.v11 import (
    MessageSegment,
    Bot,
    Event,
    MessageEvent,
    GroupMessageEvent,
)

import asyncio

import sys
from pathlib import Path
# Put `pjsk-bot/` on sys.path so `from agent...` resolves correctly.
sys.path.append(str(Path(__file__).resolve().parents[2]))
from agent.pjsk_agent import PJSKAgent
from config import get_server_base_url

from .config import Config

__plugin_meta__ = PluginMetadata(
    name="pjsk",
    description="",
    usage="",
    config=Config,
)

global_config = get_driver().config
config = Config.model_validate(global_config.model_dump())

chart = on_command("pjsk_chart")
@chart.handle()
async def handle_chart(bot: Bot, event: Event):
    # 解析参数
    command = str(event.get_message()).strip()
    if command.startswith("/pjsk_chart"):
        text = command[len("/pjsk_chart"):].strip()
    else:
        await chart.finish("命令错误")
    
    args = text.split()
    if len(args) != 2:
        await chart.finish("参数错误")
        return
    
    # 请求服务
    id = args[0]
    level = args[1]
    paramemters = {
        "id": id,
        "level": level
    }
    resp = requests.get(f"{get_server_base_url()}/pjsk/charts", params=paramemters)
    type = resp.headers['Content-Type']

    # 处理响应
    if type == "image/png":
        img = resp.content
        await chart.send(MessageSegment.image("base64://" + base64.b64encode(img).decode()))
    elif type.startswith("application/json"):
        await chart.send(resp.json()["error"])
    else:
        await chart.send("发生错误")

jacket = on_command("pjsk_jacket")
@jacket.handle()
async def handle_jacket(bot: Bot, event: Event):
    # 解析参数
    command = str(event.get_message()).strip()
    if command.startswith("/pjsk_jacket"):
        text = command[len("/pjsk_jacket"):].strip()
    else:
        await jacket.finish("命令错误")
    
    args = text.split()
    if len(args) != 1:
        await jacket.finish("参数错误")
        return
    
    # 请求服务
    id = args[0]
    paramemters = {
        "id": id
    }
    resp = requests.get(f"{get_server_base_url()}/pjsk/jackets", params=paramemters)
    type = resp.headers['Content-Type']

    # 处理响应
    if type == "image/png":
        img = resp.content
        await jacket.send(MessageSegment.image("base64://" + base64.b64encode(img).decode()))
    elif type.startswith("application/json"):
        await jacket.send(resp.json()["error"])
    else:
        await jacket.send("发生错误")

hello = on_command("pjsk_hello")
@hello.handle()
async def handle_hello():
    await hello.send("hello pjsk")


echo = on_command("pjsk_echo")
@echo.handle()
async def handle_echo(bot: Bot, event: Event):
    text = str(event.get_message()).strip()
    if text.startswith("/pjsk_echo"):
        text = text[len("/pjsk_echo"):].strip()
    args = text.split()
    arg0 = args[0] if len(args) > 0 else ""
    await echo.send(arg0)

groupe_poke = on_notice()

@groupe_poke.handle()
async def handle_poke(bot: Bot, event: Event):
    print("收到戳一戳事件")
    # 只监听机器人被戳
    if event.target_id == event.self_id:
        cpu_usage = psutil.cpu_percent(interval=1)
        cpu = str(psutil.cpu_count(logical=False)) + "/" + str(psutil.cpu_count(logical=True))
        memory_usage = psutil.virtual_memory().percent
        memory = format((psutil.virtual_memory().used / 1000000000),".3f") + "GB" + " / " + format((psutil.virtual_memory().total / 1000000000),".3f") + "GB"
        disk_usage = psutil.disk_usage('/').percent
        disk = format((psutil.disk_usage('/').used / 1000000000),".3f") + "GB" + " / " + format((psutil.disk_usage('/').total / 1000000000),".3f") + "GB"
        print("执行戳一戳响应")
        await groupe_poke.send("cpu: " + cpu + " 使用率: " + str(cpu_usage) + "%\n" +
                               "内存: " + memory + " 使用率: " + str(memory_usage) + "%\n" +
                               "磁盘: " + disk + " 使用率: " + str(disk_usage) + "%")
    else:
        print("响应未完成")

agent: PJSKAgent = PJSKAgent()

async def at_bot_rule(bot: Bot, event: Event) -> bool:
    if not isinstance(event, MessageEvent):
        return False
    if not isinstance(event, GroupMessageEvent):
        return True
    for seg in event.get_message():
        if seg.type == "at" and str(seg.data.get("qq")) == str(event.self_id):
            return True
    return False


question = on_command("chat")
@question.handle()
async def handle_question(bot: Bot, event: Event):
    text = event.get_message().extract_plain_text().strip()
    print("问题是：" + text)
    result : str | bytes = await agent.ask(text)
    if (isinstance(result,bytes)):
        await question.send(MessageSegment.image("base64://" + base64.b64encode(result).decode()))
    elif (isinstance(result,str)):
        await question.send(result)
    else:
        await question.send("模型不太聪明，自己查吧")
