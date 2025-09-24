# server_player.py
import pygame
import zmq
import time
import random
import json
import argparse

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
rep_socket = context.socket(zmq.REP)
rep_socket.bind("tcp://*:5557")    # handle connection requests

print("Server starting. You can choose your role when clients connect.")

# Parse command line arguments
parser = argparse.ArgumentParser()
parser.add_argument("--role", default=None, help="Server role: player or spectator")
args = parser.parse_args()

def ask_server_role():
    print("Server role selection:")
    print("1. Player (control a paddle)")
    print("2. Spectator (watch only)")
    choice = input("Enter 1 or 2: ").strip()
    return "player" if choice == "1" else "spectator"

# --- Connection management
player1_assigned = False  # Track if player 1 slot is taken
player2_assigned = False  # Track if player 2 slot is taken
spectators = set()        # Track spectator client IDs
last_heartbeat = {}       # Track last heartbeat from clients
player1_client_id = None  # Track which client controls player 1 (None = server)
player2_client_id = None  # Track which client controls player 2

# Get server's preferred role
if args.role:
    server_role = args.role
    print(f"Server role set to: {server_role}")
else:
    server_role = ask_server_role()

if server_role == "player":
    # Auto-assign server to Player 1 slot
    player1_assigned = True
    player1_client_id = "SERVER"
    print("Server assigned as Player 1 - Use W/S keys to control paddle")
    print(f"DEBUG: player1_assigned={player1_assigned}, player1_client_id={player1_client_id}")
    server_player_number = 1
else:
    print("Server is spectating")
    server_player_number = None

# --- Pygame setup (for server player control + display)
pygame.init()
screen = pygame.display.set_mode((WIDTH, HEIGHT))
if server_role == "player":
    pygame.display.set_caption(f"Pong Server (Player {server_player_number}) - Controls: W/S")
else:
    pygame.display.set_caption("Pong Server (Spectator)")
clock = pygame.time.Clock()
font = pygame.font.Font(None, 48)

# --- Game state
paddle_y = {1: (HEIGHT - PADDLE_HEIGHT) / 2, 2: (HEIGHT - PADDLE_HEIGHT) / 2}
scores = {1: 0, 2: 0}
ball = {"x": WIDTH/2, "y": HEIGHT/2, "vx": BALL_SPEED, "vy": 0.0}

def reset_ball(direction=1):
    ball["x"] = WIDTH/2
    ball["y"] = HEIGHT/2

    random_value = 0.2 + random.random() * 0.3
    
    # Randomly make it positive or negative
    if random.random() < 0.5:
        random_value = -random_value
        
    angle = random_value * 0.8
    ball["vx"] = BALL_SPEED * (1 if direction >= 0 else -1)
    ball["vy"] = BALL_SPEED * angle

reset_ball()

# input states
inputs = {1: {"up": False, "down": False}, 2: {"up": False, "down": False}}

def handle_connection_request():
    """Handle incoming connection requests and assign roles"""
    global player1_assigned, player2_assigned, player1_client_id, player2_client_id
    try:
        # Check for connection requests (non-blocking)
        message = rep_socket.recv_json(flags=zmq.NOBLOCK)
        if message["type"] == "connect":
            client_id = message["client_id"]
            preferred_role = message.get("role", "player")  # "player" or "spectator"
            
            if preferred_role == "player":
                # Try to assign to an available player slot
                if not player1_assigned:
                    player1_assigned = True
                    player1_client_id = client_id
                    response = {"status": "accepted", "role": "player", "player_id": 1}
                    print(f"Client {client_id} assigned as Player 1")
                elif not player2_assigned:
                    player2_assigned = True
                    player2_client_id = client_id
                    response = {"status": "accepted", "role": "player", "player_id": 2}
                    print(f"Client {client_id} assigned as Player 2")
                else:
                    # Both player slots occupied
                    spectators.add(client_id)
                    response = {"status": "accepted", "role": "spectator", "reason": "players_occupied"}
                    print(f"Client {client_id} assigned as spectator (both player slots occupied)")
            else:
                # Wants to be spectator
                spectators.add(client_id)
                response = {"status": "accepted", "role": "spectator"}
                print(f"Client {client_id} joined as spectator")
            
            last_heartbeat[client_id] = time.time()
            rep_socket.send_json(response)
    except zmq.Again:
        pass

tick = 0
last_time = time.time()
running = True
while running:
    dt = clock.tick(60) / 1000.0

    # --- handle connection requests
    handle_connection_request()

    # --- handle pygame input (server local input - only if server is a player)
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            running = False
    
    # Only process local input if server is assigned as a player and controlling a slot
    if server_role == "player":
        # Check if server is controlling Player 1
        if player1_client_id == "SERVER":
            keys = pygame.key.get_pressed()
            if keys[pygame.K_w] or keys[pygame.K_s]:  # Debug: only print when keys pressed
                print(f"Server input: W={keys[pygame.K_w]}, S={keys[pygame.K_s]}")
            inputs[1]["up"] = keys[pygame.K_w]
            inputs[1]["down"] = keys[pygame.K_s]
        # Check if server is controlling Player 2
        elif player2_client_id == "SERVER":
            keys = pygame.key.get_pressed()
            inputs[2]["up"] = keys[pygame.K_w]
            inputs[2]["down"] = keys[pygame.K_s]

    # --- handle client input
    try:
        while True:
            msg = pull_socket.recv_json(flags=zmq.NOBLOCK)
            if msg["type"] == "input":
                client_id = msg.get("client_id")
                player_id = msg.get("player")
                
                # Accept input from assigned players
                if ((player_id == 1 and client_id == player1_client_id) or 
                    (player_id == 2 and client_id == player2_client_id)):
                    inputs[player_id]["up"] = msg["up"]
                    inputs[player_id]["down"] = msg["down"]
                    last_heartbeat[client_id] = time.time()
            elif msg["type"] == "heartbeat":
                # Update heartbeat for any connected client
                client_id = msg.get("client_id")
                if client_id:
                    last_heartbeat[client_id] = time.time()
    except zmq.Again:
        pass

    # --- check for disconnected clients
    current_time = time.time()
    disconnected_clients = []
    for client_id, last_time in last_heartbeat.items():
        if current_time - last_time > 5.0:  # 5 second timeout
            disconnected_clients.append(client_id)
    
    for client_id in disconnected_clients:
        del last_heartbeat[client_id]
        if client_id == player1_client_id and client_id != "SERVER":
            print(f"Player 1 (client {client_id}) disconnected")
            player1_assigned = False
            player1_client_id = None
            inputs[1] = {"up": False, "down": False}  # Stop paddle movement
        elif client_id == player2_client_id and client_id != "SERVER":
            print(f"Player 2 (client {client_id}) disconnected")
            player2_assigned = False
            player2_client_id = None
            inputs[2] = {"up": False, "down": False}  # Stop paddle movement
        elif client_id in spectators:
            spectators.remove(client_id)
            print(f"Spectator (client {client_id}) disconnected")

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
    if LEFT_X <= ball["x"] - BALL_RADIUS <= LEFT_X + PADDLE_WIDTH:
        if paddle_y[1] <= ball["y"] <= paddle_y[1] + PADDLE_HEIGHT/2:
            ball["vx"] = abs(ball["vx"] * 1.03)
            ball["vy"] = -abs(ball["vy"])
        elif paddle_y[1] + PADDLE_HEIGHT/2 < ball["y"] <= paddle_y[1] + PADDLE_HEIGHT:
            ball["vx"] = abs(ball["vx"] * 1.03)
            ball["vy"] = abs(ball["vy"])
    if RIGHT_X <= ball["x"] + BALL_RADIUS <= RIGHT_X + PADDLE_WIDTH:
        if paddle_y[2] <= ball["y"] <= paddle_y[2] + PADDLE_HEIGHT/2:
            ball["vx"] = -abs(ball["vx"] * 1.03)
            ball["vy"] = -abs(ball["vy"])
        elif paddle_y[2] + PADDLE_HEIGHT/2 < ball["y"] <= paddle_y[2] + PADDLE_HEIGHT:
            ball["vx"] = -abs(ball["vx"] * 1.03)
            ball["vy"] = abs(ball["vy"])
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
        "scores": scores,
        "players": {
            "player1": "connected" if player1_assigned else "open",
            "player2": "connected" if player2_assigned else "open"
        },
        "spectator_count": len(spectators)
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
