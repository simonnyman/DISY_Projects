# server_player.py
import pygame
import zmq
import time
import random
import json

# --- Game constants
WIDTH, HEIGHT = 800, 600
PADDLE_WIDTH, PADDLE_HEIGHT = 10, 100
LEFT_X = 50
RIGHT_X = WIDTH - 50 - PADDLE_WIDTH
PADDLE_SPEED = 300.0
BALL_RADIUS = 8
BALL_SPEED = 320.0
TICK_RATE = 60.0

# --- ZeroMQ setup
context = zmq.Context()
pull_socket = context.socket(zmq.PULL)
pull_socket.bind("tcp://*:5555")   # receive client input
pub_socket = context.socket(zmq.PUB)
pub_socket.bind("tcp://*:5556")    # send game state
server_ip = pull_socket.getsockopt_string(zmq.LAST_ENDPOINT).split("//")[1].split(":")[0]
print("Server running (headless, logic-only mode).")


# --- Game state
paddle_y = {1: (HEIGHT - PADDLE_HEIGHT) / 2, 2: (HEIGHT - PADDLE_HEIGHT) / 2}
scores = {1: 0, 2: 0}
ball = {"x": WIDTH/2, "y": HEIGHT/2, "vx": BALL_SPEED, "vy": 0.0}

def reset_ball(direction=1):
    ball["x"] = WIDTH/2
    ball["y"] = HEIGHT/2
    angle = (random.random() - 0.5) * 0.8
    ball["vx"] = BALL_SPEED * (1 if direction >= 0 else -1)
    ball["vy"] = BALL_SPEED * angle

reset_ball()

def waiting_for_players():
    while not enough_players:
        pub_socket.send_json({"type": "waiting", "players_connected": 1, "server_ip": server_ip})

# input states
inputs = {1: {"up": False, "down": False}, 2: {"up": False, "down": False}}

last_input_time = {1: time.time(), 2: time.time()}

tick = 0

tick_rate = 1.0 / TICK_RATE
running = True
enough_players = False
while running:
    # waiting_for_players()
    start_time = time.time()
    # for last_input in last_input_time:
    #     if time.time() - last_input_time[last_input] > 5.0:
    # # --- handle client input (player 2)
    try:
        while True:
            msg = pull_socket.recv_json(flags=zmq.NOBLOCK)
            if msg["type"] == "input" and msg["player"] in (1,2):
                inputs[msg["player"]]["up"] = msg["up"]
                inputs[msg["player"]]["down"] = msg["down"]
                last_input_time[msg["player"]] = time.time()
                print(f"Received input from player {msg['player']}: up={msg['up']} down={msg['down']}")
    except zmq.Again:
        pass

    # --- update paddles
    for pid in (1,2):
        if inputs[pid]["up"]:
            paddle_y[pid] -= PADDLE_SPEED * tick_rate
        if inputs[pid]["down"]:
            paddle_y[pid] += PADDLE_SPEED * tick_rate
        paddle_y[pid] = max(0, min(HEIGHT - PADDLE_HEIGHT, paddle_y[pid]))
        last_input_time[pid] = time.time()

    # --- update ball
    ball["x"] += ball["vx"] * tick_rate
    ball["y"] += ball["vy"] * tick_rate

    if ball["y"] - BALL_RADIUS <= 0 or ball["y"] + BALL_RADIUS >= HEIGHT:
        ball["vy"] = -ball["vy"]

    # paddle collisions
    if ball["x"] - BALL_RADIUS <= LEFT_X + PADDLE_WIDTH:
        if paddle_y[1] <= ball["y"] <= paddle_y[1] + PADDLE_HEIGHT:
            ball["vx"] = abs(ball["vx"])
    if ball["x"] + BALL_RADIUS >= RIGHT_X:
        if paddle_y[2] <= ball["y"] <= paddle_y[2] + PADDLE_HEIGHT:
            ball["vx"] = -abs(ball["vx"])

    # scoring
    if ball["x"] < 0:
        scores[2] += 1
        reset_ball(direction=1)
    if ball["x"] > WIDTH:
        scores[1] += 1
        reset_ball(direction=-1) #iuerbhuifgqenirog

    # --- broadcast state
    state = {
        "type":"state",
        "tick": tick,
        "ball": ball,
        "paddles": paddle_y,
        "scores": scores
    }
    pub_socket.send_json(state)


    # Maintain tick rate
    elapsed = time.time() - start_time
    sleep_time = tick_rate - elapsed
    if sleep_time > 0:
        time.sleep(sleep_time)

    tick += 1

# TODO make server wait for players before starting the game loop
