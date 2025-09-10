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

print("Server running as Player 1.")

# --- Pygame setup (for server player control + display)
pygame.init()
screen = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Pong Server (Player 1)")
clock = pygame.time.Clock()
font = pygame.font.Font(None, 48)

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

# input states
inputs = {1: {"up": False, "down": False}, 2: {"up": False, "down": False}}

tick = 0
last_time = time.time()
running = True
while running:
    dt = clock.tick(60) / 1000.0

    # --- handle pygame input (server = player 1)
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            running = False
    keys = pygame.key.get_pressed()
    inputs[1]["up"] = keys[pygame.K_w]
    inputs[1]["down"] = keys[pygame.K_s]

    # --- handle client input (player 2)
    try:
        while True:
            msg = pull_socket.recv_json(flags=zmq.NOBLOCK)
            if msg["type"] == "input" and msg["player"] == 2:
                inputs[2]["up"] = msg["up"]
                inputs[2]["down"] = msg["down"]
    except zmq.Again:
        pass

    # --- update paddles
    for pid in (1,2):
        if inputs[pid]["up"]:
            paddle_y[pid] -= PADDLE_SPEED * dt
        if inputs[pid]["down"]:
            paddle_y[pid] += PADDLE_SPEED * dt
        paddle_y[pid] = max(0, min(HEIGHT - PADDLE_HEIGHT, paddle_y[pid]))

    # --- update ball
    ball["x"] += ball["vx"] * dt
    ball["y"] += ball["vy"] * dt

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
        reset_ball(direction=-1)

    # --- broadcast state
    state = {
        "type":"state",
        "tick": tick,
        "ball": ball,
        "paddles": paddle_y,
        "scores": scores
    }
    pub_socket.send_json(state)

    # --- render locally
    screen.fill((0,0,0))
    pygame.draw.rect(screen, (255,255,255), (LEFT_X, int(paddle_y[1]), PADDLE_WIDTH, PADDLE_HEIGHT))
    pygame.draw.rect(screen, (255,255,255), (RIGHT_X, int(paddle_y[2]), PADDLE_WIDTH, PADDLE_HEIGHT))
    pygame.draw.circle(screen, (255,255,255), (int(ball["x"]), int(ball["y"])), BALL_RADIUS)
    pygame.draw.aaline(screen, (255,255,255), (WIDTH//2, 0), (WIDTH//2, HEIGHT))
    s1 = font.render(str(scores[1]), True, (255,255,255))
    s2 = font.render(str(scores[2]), True, (255,255,255))
    screen.blit(s1, (WIDTH//4, 20))
    screen.blit(s2, (WIDTH*3//4, 20))
    pygame.display.flip()

    tick += 1

pygame.quit()
