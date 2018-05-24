#! /usr/bin/python3

from aiohttp import web
import time

async def fake_register(request):
    teamId = request.query.get("teamId")
    teamName = request.query.get("teamName")
    if teamId == "ffff" and teamName == "dirtbags":
        resp = {
            "status": "success",
            "data": None,
        }
    elif teamId and teamName:
        resp = {
            "status": "error",
            "message": "Query was correctly formed but I'm feeling cranky"
        }
    else:
        resp = {
            "status": "fail",
            "data": "You must send teamId and teamName",
        }
    return web.json_response(resp)

async def fake_state(request):
    resp = {
        "status": "success",
        "data": {
            "puzzles": {
                "sequence": [1, 2],
                "codebreaking": [10],
                "wopr": "https://appspot.com/dooted-bagel-8372/entry"
            },
            "teams": {
                "0": "Zelda",
                "1": "Defender"
            },
            "log": [
                [1526478368, "0", "sequence", 1],
                [1526478524, "1", "sequence", 1],
                [1526478536, "0", "nocode", 1]
            ],
            "notices": [
                "<a href=\"https://appspot.com/dooted-bagel-8372/entry\">WOPR category</a> is now open",
                "Event closes at 18:00 today, and will resume tomorrow at 08:00"
            ],
            "now": int(time.time()),
        }
    }
    return web.json_response(resp)

async def fake_getpuzzle(request):
    category = request.query.get("category")
    points = request.query.get("points")
    if category == "sequence" and points == "1":
        resp = {
            "status": "success",
            "data": {
                "authors": ["neale"],
                "hashes": [177627],
                "files": {
                  "happy.png": "https://cdn/assets/0904cf3a437a348bea2c49d56a3087c26a01a63c.png"
                },
                "body": "<pre><code>1 2 3 4 5 _\n</code></pre>\n",
            }
        }
    elif category and points:
        resp = {
            "status": "error",
            "message": "Query was correctly formed but I'm feeling cranky"
        }
    else:
        resp = {
            "status": "fail",
            "data": "You must send category and points"
        }
    return web.json_response(resp)

async def fake_submitanswer(request):
    teamId = request.query.get("teamId")
    category = request.query.get("category")
    points = request.query.get("points")
    answer = request.query.get("answer")
    if category == "sequence" and points == "1" and answer == "6":
        resp = {
            "status": "success",
            "data": {
                "epilog": "Now you know the answer, and knowing is half the battle. Go Joe!"
            }
        }
    elif category and points and answer:
        resp = {
            "status": "error",
            "message": "Query was correctly formed but I'm feeling cranky"
        }
    else:
        resp = {
            "status": "fail",
            "data": "You must send category and points"
        }
    return web.json_response(resp)

async def fake_submittoken(request):
    teamId = request.query.get("teamId")
    token = request.query.get("token")
    if token == "wat:30:xylep-radar-nanox":
        resp = {
            "status": "success",
            "data": {
                "category": "wat",
                "points": 30,
                "epilog": ""
            }
        }
    elif category and points and answer:
        resp = {
            "status": "error",
            "message": "Query was correctly formed but I'm feeling cranky"
        }
    else:
        resp = {
            "status": "fail",
            "data": "You must send category and points"
        }
    return web.json_response(resp)

if __name__ == "__main__":
    app = web.Application()
    app.router.add_route("GET", "/api/v3/RegisterTeam", fake_register)
    app.router.add_route("GET", "/api/v3/GetState", fake_state)
    app.router.add_route("GET", "/api/v3/GetPuzzle", fake_getpuzzle)
    app.router.add_route("GET", "/api/v3/SubmitAnswer", fake_submitanswer)
    app.router.add_route("GET", "/api/v3/SubmitToken", fake_submittoken)
    web.run_app(app)
