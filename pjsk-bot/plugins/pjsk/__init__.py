import base64
from nonebot import get_driver
from nonebot.plugin import PluginMetadata, on_command
import requests
from nonebot.adapters.onebot.v11 import MessageSegment, Bot, Event

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
    resp = requests.get(f"http://localhost:9470/pjsk/charts", params=paramemters)
    type = resp.headers['Content-Type']

    # 处理响应
    if type == "image/png":
        img = resp.content
        await chart.send(MessageSegment.image("base64://" + base64.b64encode(img).decode()))
    elif type.startswith("application/json"):
        await chart.send(resp.json()["error"])
    else:
        await chart.send("发生错误")

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