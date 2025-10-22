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

#print("Server starting. You can choose your role when clients connect.")

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
# if args.role:
#     server_role = args.role
#     print(f"Server role set to: {server_role}")
# else:
#     server_role = ask_server_role()

# if server_role == "player":
#     # Auto-assign server to Player 1 slot
#     player1_assigned = True
#     player1_client_id = "SERVER"
#     print("Server assigned as Player 1 - Use W/S keys to control paddle")
#     print(f"DEBUG: player1_assigned={player1_assigned}, player1_client_id={player1_client_id}")
#     server_player_number = 1
# else:
#     print("Server is spectating")
#     server_player_number = None

# --- Pygame setup (for server player control + display)
# pygame.init()
# screen = pygame.display.set_mode((WIDTH, HEIGHT))
# if server_role == "player":
#     pygame.display.set_caption(f"Pong Server (Player {server_player_number}) - Controls: W/S")
# else:
#     pygame.display.set_caption("Pong Server (Spectator)")
clock = pygame.time.Clock()
# font = pygame.font.Font(None, 48)

# --- Game state
paddle_y = {1: (HEIGHT - PADDLE_HEIGHT) / 2, 2: (HEIGHT - PADDLE_HEIGHT) / 2}
scores = {1: 0, 2: 0}
ball = {"x": WIDTH/2, "y": HEIGHT/2, "vx": BALL_SPEED, "vy": 0.0}
paused = False

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

# def waiting_for_players():
#     while not enough_players:
#         pub_socket.send_json({"type": "waiting", "players_connected": 1, "server_ip": server_ip})

# input states
inputs = {1: {"up": False, "down": False}, 2: {"up": False, "down": False}}

#last_input_time = {1: time.time(), 2: time.time()}

def handle_role_update(message):
    """Handle role update requests from clients"""
    global player1_assigned, player2_assigned, player1_client_id, player2_client_id, spectators
    client_id = message.get("client_id")
    new_role = message.get("role")
    if not client_id or not new_role:
        rep_socket.send_json({"status": "error", "reason": "invalid_request"})
        return

    # Helper to free a player slot if this client currently occupies it
    def free_if_assigned(cid):
        global player1_assigned, player2_assigned, player1_client_id, player2_client_id
        if player1_client_id == cid:
            player1_assigned = False
            player1_client_id = None
            #print(f"Freed Player 1 slot previously held by {cid}")
        if player2_client_id == cid:
            player2_assigned = False
            player2_client_id = None
            #print(f"Freed Player 2 slot previously held by {cid}")

    # If client asks to be a player
    if new_role == "player":
        # If already a player, confirm their slot
        if client_id == player1_client_id:
            response = {"status": "accepted", "role": "player", "player_id": 1, "client_id": client_id}
            #print(f"Client {client_id} remains Player 1 via role update")
        elif client_id == player2_client_id:
            response = {"status": "accepted", "role": "player", "player_id": 2, "client_id": client_id}
            #print(f"Client {client_id} remains Player 2 via role update")
        else:
            # Try to assign to an available player slot
            if not player1_assigned:
                player1_assigned = True
                player1_client_id = client_id
                spectators.discard(client_id)
                response = {"status": "accepted", "role": "player", "player_id": 1, "client_id": client_id}
                #print(f"Client {client_id} assigned as Player 1 via role update")
            elif not player2_assigned:
                player2_assigned = True
                player2_client_id = client_id
                spectators.discard(client_id)
                response = {"status": "accepted", "role": "player", "player_id": 2, "client_id": client_id}
                #print(f"Client {client_id} assigned as Player 2 via role update")
            else:
                # Both player slots occupied
                spectators.add(client_id)
                response = {"status": "accepted", "role": "spectator", "reason": "players_occupied", "client_id": client_id}
                #print(f"Client {client_id} assigned as spectator (both player slots occupied) via role update")

    else:
        # Wants to be spectator: free any player slot they might be holding
        free_if_assigned(client_id)
        spectators.add(client_id)
        response = {"status": "accepted", "role": "spectator", "client_id": client_id}
        #print(f"Client {client_id} joined as spectator via role update")

    last_heartbeat[client_id] = time.time()
    rep_socket.send_json(response)
    #print("Sent response:", response)

def handle_connection_request():
    """Handle incoming connection requests and assign roles"""
    global player1_assigned, player2_assigned, player1_client_id, player2_client_id
    try:
        # Check for connection requests (non-blocking)
        #print("Checking for connection requests...")
        message = rep_socket.recv_json(flags=zmq.NOBLOCK)
        if message["type"] == "role_update":
            handle_role_update(message)
            return
        if message["type"] == "connect":
            #print("Received connection request:", message)
            client_id = message["client_id"]
            preferred_role = message.get("role", "player")  # "player" or "spectator"
            
            if preferred_role == "player":
                # Try to assign to an available player slot
                if not player1_assigned:
                    player1_assigned = True
                    player1_client_id = client_id
                    response = {"status": "accepted", "role": "player", "player_id": 1, "client_id": client_id}
                    #print(f"Client {client_id} assigned as Player 1")
                elif not player2_assigned:
                    player2_assigned = True
                    player2_client_id = client_id
                    response = {"status": "accepted", "role": "player", "player_id": 2, "client_id": client_id}
                    #print(f"Client {client_id} assigned as Player 2")
                else:
                    # Both player slots occupied
                    spectators.add(client_id)
                    response = {"status": "accepted", "role": "spectator", "reason": "players_occupied", "client_id": client_id}
                    #print(f"Client {client_id} assigned as spectator (both player slots occupied)")
            else:
                # Wants to be spectator
                spectators.add(client_id)
                response = {"status": "accepted", "role": "spectator", "client_id": client_id}
                #print(f"Client {client_id} joined as spectator")
            
            last_heartbeat[client_id] = time.time()
            rep_socket.send_json(response)
            #print("Sent response:", response)
        else:
            rep_socket.send_json({"status": "error", "reason": "invalid_request"})
    except zmq.Again:
        pass

tick = 0

tick_rate = 1.0 / TICK_RATE
running = True
enough_players = False
last_time = time.time()

# --- Wait for two players to connect before starting the game loop
#print("Waiting for 2 players to connect before starting the game...")
wait_start = time.time()
while not (player1_assigned and player2_assigned) and running:
    # Process any incoming input/heartbeats so clients can register
    try:
        while True:
            msg = pull_socket.recv_json(flags=zmq.NOBLOCK)
            if msg["type"] == "input":
                # ignore inputs until players are assigned
                client_id = msg.get("client_id")
                last_heartbeat[client_id] = time.time()
            elif msg["type"] == "heartbeat":
                client_id = msg.get("client_id")
                if client_id:
                    last_heartbeat[client_id] = time.time()
    except zmq.Again:
        pass

    # Handle any pending connection/role requests
    try:
        handle_connection_request()
    except Exception as e:
        print("Error while handling connection request in wait loop:", e)

    # Clean up timed-out clients while waiting
    current_time = time.time()
    disconnected_clients = []
    for client_id, last_time_hb in list(last_heartbeat.items()):
        if current_time - last_time_hb > 5.0:
            disconnected_clients.append(client_id)
    for client_id in disconnected_clients:
        del last_heartbeat[client_id]
        if client_id == player1_client_id and client_id != "SERVER":
            #print(f"Player 1 (client {client_id}) disconnected while waiting")
            player1_assigned = False
            player1_client_id = None
        elif client_id == player2_client_id and client_id != "SERVER":
            #print(f"Player 2 (client {client_id}) disconnected while waiting")
            player2_assigned = False
            player2_client_id = None
        elif client_id in spectators:
            spectators.remove(client_id)
    
    # Informative periodic status and broadcast waiting state so clients can display it
    if int(time.time() - wait_start) % 1 == 0:
        p1 = 'connected' if player1_assigned else 'open'
        p2 = 'connected' if player2_assigned else 'open'
        # print once per second (will naturally duplicate in console but it's fine)
        #print(f"Waiting status: player1={p1}, player2={p2}, spectators={len(spectators)}")
        waiting_state = {
            "type": "waiting",
            "message": "Waiting for players",
            "players": {"player1": p1, "player2": p2},
            "spectator_count": len(spectators)
        }
        try:
            pub_socket.send_json(waiting_state)
        except Exception:
            # ignore pub errors while waiting
            pass

    time.sleep(0.1)

#print("Both players connected. Starting game loop.")

while running:
    dt = clock.tick(60) / 1000.0

    # --- handle connection requests
    #handle_connection_request()

    # --- handle pygame input (server local input - only if server is a player)
    # for event in pygame.event.get():
    #     if event.type == pygame.QUIT:
    #         running = False
    
    # Only process local input if server is assigned as a player and controlling a slot
    # if server_role == "player":
    #     # Check if server is controlling Player 1
    #     if player1_client_id == "SERVER":
    #         keys = pygame.key.get_pressed()
    #         if keys[pygame.K_w] or keys[pygame.K_s]:  # Debug: only print when keys pressed
    #             print(f"Server input: W={keys[pygame.K_w]}, S={keys[pygame.K_s]}")
    #         inputs[1]["up"] = keys[pygame.K_w]
    #         inputs[1]["down"] = keys[pygame.K_s]
    #     # Check if server is controlling Player 2
    #     elif player2_client_id == "SERVER":
    #         keys = pygame.key.get_pressed()
    #         inputs[2]["up"] = keys[pygame.K_w]
    #         inputs[2]["down"] = keys[pygame.K_s]

    # --- handle client input
    try:
        while True:
            msg = pull_socket.recv_json(flags=zmq.NOBLOCK)
            if msg["type"] == "input":
                client_id = msg.get("client_id")
                player_id = msg.get("player")
                #print("Received input message:", msg)
                # Accept input from assigned players
                if ((player_id == 1 and client_id == player1_client_id) or 
                    (player_id == 2 and client_id == player2_client_id)):
                    #print("Received input message:", msg)
                    inputs[player_id]["up"] = msg["up"]
                    inputs[player_id]["down"] = msg["down"]
                    last_heartbeat[client_id] = time.time()
                    #print(f"Received input from Player {player_id} (client {client_id}): up={msg['up']} down={msg['down']}")
            elif msg["type"] == "pause":
                # pause/resume requested by a client - only accept from assigned players
                client_id = msg.get("client_id")
                action = msg.get("action")
                is_player = (client_id == player1_client_id) or (client_id == player2_client_id)
                if not is_player:
                    print(f"Ignored pause request from non-player client {client_id}")
                else:
                    if action == "pause":
                        paused = True
                        print(f"Client {client_id} (player) requested PAUSE")
                    elif action == "resume":
                        paused = False
                        print(f"Client {client_id} (player) requested RESUME")
            elif msg["type"] == "heartbeat":
                # Update heartbeat for any connected client
                client_id = msg.get("client_id")
                if client_id:
                    last_heartbeat[client_id] = time.time()
    except zmq.Again:
        pass
    # --- handle new connection requests
    try:
        handle_connection_request()
    except Exception as e:
        print("Error handling connection request:", e)
    
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

    # --- update paddles and ball (skip updates if paused)
    if not paused:
        for pid in (1,2):
            if inputs[pid]["up"]:
                paddle_y[pid] -= PADDLE_SPEED * tick_rate
            if inputs[pid]["down"]:
                paddle_y[pid] += PADDLE_SPEED * tick_rate
            paddle_y[pid] = max(0, min(HEIGHT - PADDLE_HEIGHT, paddle_y[pid]))

        # --- update ball
        ball["x"] += ball["vx"] * tick_rate
        ball["y"] += ball["vy"] * tick_rate

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
        "paused": paused,
        "players": {
            "player1": "connected" if player1_assigned else "open",
            "player2": "connected" if player2_assigned else "open"
        },
        "spectator_count": len(spectators)
    }
    pub_socket.send_json(state)


    # Maintain tick rate
    elapsed = time.time() - last_time
    sleep_time = tick_rate - elapsed
    if sleep_time > 0:
        time.sleep(sleep_time)
    last_time = time.time()

    tick += 1

# TODO make server wait for players before starting the game loop
