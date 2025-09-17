
# client_player.py
import pygame
import zmq
import argparse
import subprocess
import sys
import time

def ask_mode():
    print("Select mode:")
    print("1. Connect to existing server")
    print("2. Start a new server on this machine")
    choice = input("Enter 1 or 2: ").strip()
    return choice

parser = argparse.ArgumentParser()
parser.add_argument("--server", default=None, help="Server IP")
args = parser.parse_args()

server_ip = None
server_process = None

if args.server:
    server_ip = args.server
else:
    mode = ask_mode()
    if mode == "2":
        # Start server as subprocess
        print("Starting server locally...")
        server_process = subprocess.Popen([sys.executable, "server_zmq.py"])
        server_ip = "127.0.0.1"
        # Wait a moment for server to start
        time.sleep(1)
    else:
        ip = input("Enter server IP (default 127.0.0.1): ").strip()
        server_ip = ip if ip else "127.0.0.1"

# --- ZeroMQ setup
context = zmq.Context()
push_socket = context.socket(zmq.PUSH)
push_socket.connect(f"tcp://{server_ip}:5555")
sub_socket = context.socket(zmq.SUB)
sub_socket.connect(f"tcp://{server_ip}:5556")
sub_socket.setsockopt_string(zmq.SUBSCRIBE, "")

# --- Pygame setup
WIDTH, HEIGHT = 800, 600
PADDLE_WIDTH, PADDLE_HEIGHT = 10, 100
LEFT_X = 50
RIGHT_X = WIDTH - 50 - PADDLE_WIDTH
BALL_RADIUS = 8

pygame.init()
screen = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Pong Client (Player 2)")
clock = pygame.time.Clock()
font = pygame.font.Font(None, 48)

remote_state = None
up, down = False, False
running = True
while running:
    dt = clock.tick(60) / 1000.0

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
    keys = pygame.key.get_pressed()
    up, down = keys[pygame.K_UP], keys[pygame.K_DOWN]

    push_socket.send_json({"type":"input","player":2,"up":up,"down":down})

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

        label = font.render("You: Player 2", True, (200,200,200))
        screen.blit(label, (10, HEIGHT-50))
    else:
        text = font.render("Waiting for server...", True, (255,255,255))
        screen.blit(text, (WIDTH//4, HEIGHT//2))

    pygame.display.flip()


pygame.quit()
if server_process:
    print("Shutting down local server...")
    server_process.terminate()
