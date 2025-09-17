
# client_player.py
from multiprocessing import context
import pygame
import zmq
import argparse
import subprocess
import sys
import time

def start_server():
    # Start server as subprocess
    print("Starting server locally...")
    server_process = subprocess.Popen([sys.executable, "server_zmq.py"])
    server_ip = "127.0.0.1"
    # Wait a moment for server to start
    time.sleep(1)
    return server_ip, server_process

# --- ZeroMQ setup
def connect_locally(server_ip):
    context = zmq.Context()
    push_socket = context.socket(zmq.PUSH)
    push_socket.connect(f"tcp://{server_ip}:5555")
    sub_socket = context.socket(zmq.SUB)
    sub_socket.connect(f"tcp://{server_ip}:5556")
    sub_socket.setsockopt_string(zmq.SUBSCRIBE, "")

    return sub_socket , push_socket


def connect_to_server(server_ip):
    context = zmq.Context()
    push_socket = context.socket(zmq.PUSH)
    push_socket.connect(f"tcp://{server_ip}:5555")
    sub_socket = context.socket(zmq.SUB)
    sub_socket.connect(f"tcp://{server_ip}:5556")
    sub_socket.setsockopt_string(zmq.SUBSCRIBE, "")

    return sub_socket , push_socket
    


# --- Pygame setup
WIDTH, HEIGHT = 800, 600
PADDLE_WIDTH, PADDLE_HEIGHT = 10, 100
LEFT_X = 50
RIGHT_X = WIDTH - 50 - PADDLE_WIDTH
BALL_RADIUS = 8

pygame.init()
screen = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Pong Client")
clock = pygame.time.Clock()
font = pygame.font.Font(None, 48)


# Button for starting server
buttonServer_rect = pygame.Rect(WIDTH//2 - 100, HEIGHT//2 + 40, 200, 50)
buttonServer_color = (70, 130, 180)
server_text = font.render("Start Server", True, (255, 255, 255))
# pygame.draw.rect(screen, buttonServer_color, buttonServer_rect)
# screen.blit(server_text, (buttonServer_rect.x + 10, buttonServer_rect.y + 10))

# Button for connecting to server
buttonConnect_rect = pygame.Rect(WIDTH//2 - 100, HEIGHT//2 - 60, 200, 50)
buttonConnect_color = (70, 130, 180)
connect_text = font.render("Connect to Server", True, (255, 255, 255))
# pygame.draw.rect(screen, buttonConnect_color, buttonConnect_rect)
# screen.blit(connect_text, (buttonConnect_rect.x + 10, buttonConnect_rect.y + 10))

remote_state = None
up, down = False, False
running = True
waiting = True


# Input for server IP when connecting
text_input_active = False
ip_input = ""
input_box_rect = pygame.Rect(WIDTH//2 - 100, HEIGHT//2 + 60, 200, 40)
input_box_color = (50, 50, 50)
input_text_color = (255, 255, 255)
input_border_color = (200, 200, 200)
player_num = 2  # default
server_process = None
while running:
    while waiting:
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                running = False
                waiting = False
            elif event.type == pygame.MOUSEBUTTONDOWN:
                if not text_input_active:
                    if buttonServer_rect.collidepoint(event.pos):
                        server_ip, server_process = start_server()
                        sub_socket, push_socket = connect_locally(server_ip)
                        player_num = 1
                        waiting = False
                    elif buttonConnect_rect.collidepoint(event.pos):
                        text_input_active = True
                # If text_input_active and input_box clicked, do nothing (keep active)
            elif event.type == pygame.KEYDOWN and text_input_active:
                if event.key == pygame.K_RETURN:
                    server_ip = ip_input.strip() if ip_input.strip() else "127.0.0.1"
                    sub_socket, push_socket = connect_to_server(server_ip)
                    player_num = 2
                    waiting = False
                    text_input_active = False
                elif event.key == pygame.K_BACKSPACE:
                    ip_input = ip_input[:-1]
                else:
                    if len(ip_input) < 30 and (event.unicode.isdigit() or event.unicode == '.' or event.unicode == ':'):
                        ip_input += event.unicode
        # Draw menu every frame
        screen.fill((0,0,0))
        if text_input_active:
            pygame.draw.rect(screen, input_box_color, input_box_rect)
            pygame.draw.rect(screen, input_border_color, input_box_rect, 2)
            ip_surf = font.render(ip_input, True, input_text_color)
            screen.blit(ip_surf, (input_box_rect.x + 5, input_box_rect.y + 5))
        else:
            pygame.draw.rect(screen, buttonServer_color, buttonServer_rect)
            screen.blit(server_text, (buttonServer_rect.x + 10, buttonServer_rect.y + 10))
            pygame.draw.rect(screen, buttonConnect_color, buttonConnect_rect)
            screen.blit(connect_text, (buttonConnect_rect.x + 10, buttonConnect_rect.y + 10))

        pygame.display.flip()
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

    push_socket.send_json({"type":"input","player":player_num,"up":up,"down":down})

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

        label = font.render(f"You: Player {player_num}", True, (200,200,200))
        screen.blit(label, (10, HEIGHT-50))
    else:
        text = font.render("Waiting for server...", True, (255,255,255))
        screen.blit(text, (WIDTH//4, HEIGHT//2))

    pygame.display.flip()


pygame.quit()
if server_process:
    print("Shutting down local server...")
    server_process.terminate()
