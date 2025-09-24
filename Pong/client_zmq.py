# client_player.py
import pygame
import zmq
import argparse
import subprocess
import sys
import time
import uuid

def ask_mode():
    print("Select mode:")
    print("1. Connect to existing server")
    print("2. Start a new server on this machine")
    choice = input("Enter 1 or 2: ").strip()
    return choice

def ask_role():
    print("Select your role:")
    print("1. Player (control a paddle)")
    print("2. Spectator (watch only)")
    choice = input("Enter 1 or 2: ").strip()
    return "player" if choice == "1" else "spectator"

parser = argparse.ArgumentParser()
parser.add_argument("--server", default=None, help="Server IP")
args = parser.parse_args()

server_ip = None
server_process = None
preferred_role = "player"

if args.server:
    server_ip = args.server
    preferred_role = ask_role()
else:
    mode = ask_mode()
    if mode == "2":
        # Ask what role they want first
        host_role = ask_role()
        
        # Start server as subprocess with the chosen role
        print("Starting server locally...")
        server_process = subprocess.Popen([sys.executable, "server_zmq.py", "--role", host_role])
        server_ip = "127.0.0.1"
        # Wait a moment for server to start
        time.sleep(2)
        
        # Always exit after starting server - server window handles everything
        if host_role == "player":
            print("Server is running with you as a player. Use the server window to play.")
        else:
            print("Server is running with you as spectator. Use the server window to watch.")
        print("Close the server window to stop the game.")
        sys.exit(0)
    else:
        ip = input("Enter server IP (default 127.0.0.1): ").strip()
        server_ip = ip if ip else "127.0.0.1"
        preferred_role = ask_role()

# --- ZeroMQ setup
context = zmq.Context()
push_socket = context.socket(zmq.PUSH)
push_socket.connect(f"tcp://{server_ip}:5555")
sub_socket = context.socket(zmq.SUB)
sub_socket.connect(f"tcp://{server_ip}:5556")
sub_socket.setsockopt_string(zmq.SUBSCRIBE, "")
req_socket = context.socket(zmq.REQ)
req_socket.connect(f"tcp://{server_ip}:5557")

# --- Connection negotiation
client_id = str(uuid.uuid4())
connect_request = {
    "type": "connect",
    "client_id": client_id,
    "role": preferred_role
}
req_socket.send_json(connect_request)
connection_response = req_socket.recv_json()

if connection_response["status"] != "accepted":
    print("Connection rejected by server")
    sys.exit(1)

assigned_role = connection_response["role"]
player_id = connection_response.get("player_id", None)
reason = connection_response.get("reason", None)

print(f"Connected as {assigned_role}")
if assigned_role == "player":
    print(f"You are Player {player_id}")
elif reason == "players_occupied":
    print("You are spectating (both player slots are occupied)")
else:
    print("You are spectating the game")

# --- Pygame setup
WIDTH, HEIGHT = 800, 600
PADDLE_WIDTH, PADDLE_HEIGHT = 10, 100
LEFT_X = 50
RIGHT_X = WIDTH - 50 - PADDLE_WIDTH
BALL_RADIUS = 8

pygame.init()
screen = pygame.display.set_mode((WIDTH, HEIGHT))
if assigned_role == "spectator":
    pygame.display.set_caption("Pong Spectator")
else:
    pygame.display.set_caption(f"Pong Client (Player {player_id})")
clock = pygame.time.Clock()
font = pygame.font.Font(None, 48)

remote_state = None
up, down = False, False
running = True
last_heartbeat = time.time()

while running:
    dt = clock.tick(60) / 1000.0
    current_time = time.time()

    # Send heartbeat every 2 seconds
    if current_time - last_heartbeat > 2.0:
        push_socket.send_json({"type": "heartbeat", "client_id": client_id})
        last_heartbeat = current_time

    # --- recv state from server
    try:
        while True:
            msg = sub_socket.recv_json(flags=zmq.NOBLOCK)
            if msg["type"] == "state":
                remote_state = msg
    except zmq.Again:
        pass

    # --- handle input
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            running = False
    
    # Only handle input if we're a player
    if assigned_role == "player":
        keys = pygame.key.get_pressed()
        # Use different controls for Player 1 vs Player 2
        if player_id == 1:
            up, down = keys[pygame.K_w], keys[pygame.K_s]
        else:  # player_id == 2
            up, down = keys[pygame.K_UP], keys[pygame.K_DOWN]
        
        push_socket.send_json({
            "type": "input",
            "client_id": client_id,
            "player": player_id,
            "up": up,
            "down": down
        })

    # --- render
    screen.fill((0,0,0))
    if remote_state:
        b = remote_state["ball"]
        p1y = remote_state["paddles"]["1"]
        p2y = remote_state["paddles"]["2"]
        s1 = remote_state["scores"]["1"]
        s2 = remote_state["scores"]["2"]

        pygame.draw.rect(screen, (255,255,255), (LEFT_X, int(p1y), PADDLE_WIDTH, PADDLE_HEIGHT))
        pygame.draw.rect(screen, (255,255,255), (RIGHT_X, int(p2y), PADDLE_WIDTH, PADDLE_HEIGHT))
        pygame.draw.circle(screen, (255,255,255), (int(b["x"]), int(b["y"])), BALL_RADIUS)
        pygame.draw.aaline(screen, (255,255,255), (WIDTH//2, 0), (WIDTH//2, HEIGHT))

        s1_surf = font.render(str(s1), True, (255,255,255))
        s2_surf = font.render(str(s2), True, (255,255,255))
        screen.blit(s1_surf, (WIDTH//4, 20))
        screen.blit(s2_surf, (WIDTH*3//4, 20))

        # Show role and connection info
        if assigned_role == "spectator":
            label = font.render("SPECTATOR", True, (200,200,100))
            screen.blit(label, (10, HEIGHT-50))
            
            # Show connection status
            player1_status = remote_state.get("players", {}).get("player1", "unknown")
            player2_status = remote_state.get("players", {}).get("player2", "unknown")
            spectator_count = remote_state.get("spectator_count", 0)
            status_text = f"P1: {player1_status} | P2: {player2_status} | Spectators: {spectator_count}"
            status_surf = pygame.font.Font(None, 24).render(status_text, True, (150,150,150))
            screen.blit(status_surf, (10, HEIGHT-25))
        else:
            label = font.render(f"You: Player {player_id}", True, (200,200,200))
            screen.blit(label, (10, HEIGHT-50))
            
            # Show controls
            if player_id == 1:
                controls_text = "Controls: W/S"
            else:
                controls_text = "Controls: UP/DOWN"
            controls_surf = pygame.font.Font(None, 24).render(controls_text, True, (150,150,150))
            screen.blit(controls_surf, (10, HEIGHT-25))
    else:
        text = font.render("Waiting for server...", True, (255,255,255))
        screen.blit(text, (WIDTH//4, HEIGHT//2))

    pygame.display.flip()


pygame.quit()
if server_process:
    print("Shutting down local server...")
    server_process.terminate()
