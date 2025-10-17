# client_player.py
from multiprocessing import context
import pygame
import zmq
import argparse
import subprocess
import sys
import time
import uuid

def cleanup_and_exit():
    """Close sockets, shut down server (if local), and quit pygame cleanly."""
    try:
        pygame.quit()
    except Exception:
        pass

    # Close sockets if they exist
    for s in ("sub_socket", "push_socket", "req_socket"):
        try:
            globals()[s].close(0)
        except Exception:
            pass

    # Stop local server subprocess if one was started
    try:
        if server_process:
            print("Shutting down local server...")
            server_process.terminate()
    except Exception:
        pass

    sys.exit(0)

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
assigned_role = "player"
client_id = str(uuid.uuid4())[:8]
#client negotiation
connection_request= {
    "type": "connect",
    "client_id": client_id,
    "role": preferred_role
}


player_id = 1

# if args.server:
#     server_ip = args.server
#     assigned_role = ask_role()
# else:
#     mode = ask_mode()
#     if mode == "2":
#         # Ask what role they want first
#         host_role = ask_role()
        
#         # Start server as subprocess with the chosen role
#         print("Starting server locally...")
#         server_process = subprocess.Popen([sys.executable, "server_zmq.py", "--role", host_role])
#         server_ip = "127.0.0.1"
#         # Wait a moment for server to start
#         time.sleep(2)
        
#         # Always exit after starting server - server window handles everything
#         if host_role == "player":
#             print("Server is running with you as a player. Use the server window to play.")
#         else:
#             print("Server is running with you as spectator. Use the server window to watch.")
#         print("Close the server window to stop the game.")
#         sys.exit(0)
#     else:
#         ip = input("Enter server IP (default 127.0.0.1): ").strip()
#         server_ip = ip if ip else "127.0.0.1"
#         preferred_role = ask_role()
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
    req_socket = context.socket(zmq.REQ)
    req_socket.connect(f"tcp://{server_ip}:5557")
    print("Connected to local server.")

    return sub_socket , push_socket , req_socket


def connect_to_server(server_ip):
    context = zmq.Context()
    push_socket = context.socket(zmq.PUSH)
    push_socket.connect(f"tcp://{server_ip}:5555")
    sub_socket = context.socket(zmq.SUB)
    sub_socket.connect(f"tcp://{server_ip}:5556")
    sub_socket.setsockopt_string(zmq.SUBSCRIBE, "")
    req_socket = context.socket(zmq.REQ)
    req_socket.connect(f"tcp://{server_ip}:5557")

    return sub_socket , push_socket , req_socket


def connect_client_id():
    #print(f"Sent connection request as {preferred_role} with client_id {client_id}")
    # get preferred role
    preferred_role = post_connection_menu()
    connection_request["role"] = preferred_role
    req_socket.send_json(connection_request)
    # Wait for server response
    try:
        print("Waiting for server response...")
        poller = zmq.Poller()
        poller.register(req_socket, zmq.POLLIN)
        socks = dict(poller.poll(5000))  # wait up to 5 seconds
        if req_socket not in socks:
            print("No response from server within timeout. Exiting.")
            return None, None
        response = req_socket.recv_json(flags=zmq.NOBLOCK)
        print("Received response:", response)
        if response["status"] == "accepted" and response["client_id"] == client_id and response["role"] == "player":
            assigned_role = response["role"]
            player_id = response.get("player_id", None)
            print(f"Connected to server as {assigned_role}.")
            if assigned_role == "player":
                print(f"You are Player {player_id}.")
            #req_socket.close()  # Close the REQ socket after use
            return assigned_role, player_id
        elif response["status"] == "accepted" and response["client_id"] == client_id and response["role"] == "spectator":
            assigned_role = response["role"]
            print(f"Connected to server as {assigned_role}.")
            #req_socket.close()  # Close the REQ socket after use
            return assigned_role, None
        else:
            print("Unexpected response from server:", response)
            return None, None
    except zmq.Again:
        print("No response from server. Exiting.")
        return None, None

# def update_role(new_role):
#     global assigned_role
#     assigned_role = new_role
#     pygame.display.set_caption(f"Pong Client (Role: {assigned_role})")


def post_connection_menu():
    menu_running = True
    play_rect = pygame.Rect(WIDTH//2 - 100, HEIGHT//2 - 30, 200, 50)
    spectate_rect = pygame.Rect(WIDTH//2 - 100, HEIGHT//2 + 40, 200, 50)
    play_text = font.render("Play", True, (255,255,255))
    spectate_text = font.render("Spectate", True, (255,255,255))
    chosen_role = None
    while menu_running:
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                pygame.quit()
                sys.exit(0)
            elif event.type == pygame.MOUSEBUTTONDOWN:
                if play_rect.collidepoint(event.pos):
                    chosen_role = "player"
                    #update_role(chosen_role)
                    menu_running = False
                elif spectate_rect.collidepoint(event.pos):
                    chosen_role = "spectator"
                    #update_role(chosen_role)
                    menu_running = False
        screen.fill((0,0,0))
        pygame.draw.rect(screen, (70,130,180), play_rect)
        screen.blit(play_text, (play_rect.x + 40, play_rect.y + 10))
        pygame.draw.rect(screen, (70,130,180), spectate_rect)
        screen.blit(spectate_text, (spectate_rect.x + 20, spectate_rect.y + 10))
        pygame.display.flip()
    return chosen_role

def update_role(new_role):
    global assigned_role
    assigned_role = new_role
    pygame.display.set_caption(f"Pong Client (Role: {assigned_role})")
    msg = {"type": "role_update", "role": assigned_role}
    req_socket.send_json(msg)


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
    pygame.display.set_caption(f"Pong Client (Player {assigned_role})")
clock = pygame.time.Clock()
font = pygame.font.Font(None, 48)

# UI helpers for pause menu
def ui_draw_panel(screen, rect):
    """Draws a semi-transparent dark panel with rounded corners."""
    surf = pygame.Surface(rect.size, pygame.SRCALPHA)
    pygame.draw.rect(surf, (20, 20, 20, 235), surf.get_rect(), border_radius=16)
    pygame.draw.rect(surf, (255, 255, 255, 25), surf.get_rect(), width=2, border_radius=16)
    screen.blit(surf, rect.topleft)

def ui_draw_btn(screen, rect, text, font_hint_size=28, hot=False):
    """Draws a button with optional highlight."""
    font_local = pygame.font.SysFont(None, font_hint_size)
    base = (70,130,180) if not hot else (90,160,210)
    surf = pygame.Surface(rect.size, pygame.SRCALPHA)
    pygame.draw.rect(surf, base, surf.get_rect(), border_radius=12)
    pygame.draw.rect(surf, (255,255,255,40), surf.get_rect(), width=2, border_radius=12)
    txt = font_local.render(text, True, (255,255,255))
    surf.blit(txt, txt.get_rect(center=surf.get_rect().center))
    screen.blit(surf, rect.topleft)

# Pause menu (blocking)
def show_pause_menu(screen, push_socket, client_id, hb_interval=2.0):
    """
    Displays a blocking pause menu with 'Resume' and 'Quit'.
    Keeps sending heartbeat to server so you don't get disconnected.
    """
    W, H = screen.get_size()
    font_big = pygame.font.SysFont(None, 60)

    # Button layout
    bw, bh = 200, 50
    panel = pygame.Rect(0, 0, int(W*0.6), int(H*0.4))
    panel.center = (W//2, H//2)
    cx, cy = panel.center
    btn_resume = pygame.Rect(cx - bw//2, cy - bh - 5, bw, bh)
    btn_quit   = pygame.Rect(cx - bw//2, cy + 5,       bw, bh)

    clock = pygame.time.Clock()
    selected = 0
    last_hb = 0.0

    while True:
        # Send heartbeat while paused
        now = pygame.time.get_ticks() / 1000.0
        if push_socket is not None and (now - last_hb) >= hb_interval:
            try:
                push_socket.send_json({"type":"heartbeat","client_id":client_id}, flags=zmq.NOBLOCK)
            except Exception:
                pass
            last_hb = now

        for e in pygame.event.get():
            if e.type == pygame.QUIT:
                return "quit"

            if e.type == pygame.KEYDOWN:
                if e.key in (pygame.K_ESCAPE, pygame.K_p): return "resume"
                if e.key in (pygame.K_UP, pygame.K_w):     selected = (selected - 1) % 2
                if e.key in (pygame.K_DOWN, pygame.K_s):   selected = (selected + 1) % 2
                if e.key in (pygame.K_RETURN, pygame.K_SPACE): return "resume" if selected==0 else "quit"
                if e.key == pygame.K_q: return "quit"
            if e.type == pygame.MOUSEBUTTONDOWN:
                if btn_resume.collidepoint(e.pos): return "resume"
                if btn_quit.collidepoint(e.pos):   return "quit"

        # Mouse hover updates selection
        mx, my = pygame.mouse.get_pos()
        if btn_resume.collidepoint((mx,my)): selected = 0
        elif btn_quit.collidepoint((mx,my)): selected = 1

        # Draw overlay
        dim = pygame.Surface((W, H), pygame.SRCALPHA)
        dim.fill((0,0,0,170))
        screen.blit(dim, (0,0))
        ui_draw_panel(screen, panel)

        title = font_big.render("Paused", True, (240,240,240))
        screen.blit(title, title.get_rect(midtop=(panel.centerx, panel.y+18)))

        ui_draw_btn(screen, btn_resume, "Resume (Esc/P)", 28, hot=(selected==0))
        ui_draw_btn(screen, btn_quit,   "Quit Game (Q)",  28, hot=(selected==1))

        pygame.display.flip()
        clock.tick(60)

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
last_heartbeat = time.time()

waiting = True


# Input for server IP when connecting
text_input_active = False
ip_input = ""
input_box_rect = pygame.Rect(WIDTH//2 - 100, HEIGHT//2 + 60, 200, 40)
input_box_color = (50, 50, 50)
input_text_color = (255, 255, 255)
input_border_color = (200, 200, 200)
#player_num = 2  # default
server_process = None
while running:
    while waiting:
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                cleanup_and_exit()
            elif event.type == pygame.MOUSEBUTTONDOWN:
                if not text_input_active:
                    if buttonServer_rect.collidepoint(event.pos):
                        server_ip, server_process = start_server()
                        #time.sleep(1)
                        sub_socket, push_socket , req_socket = connect_locally(server_ip)
                        assigned_role, player_id = connect_client_id()
                        #player_num = 1
                        waiting = False
                    elif buttonConnect_rect.collidepoint(event.pos):
                        text_input_active = True
                # If text_input_active and input_box clicked, do nothing (keep active)
            elif event.type == pygame.KEYDOWN and text_input_active:
                if event.key == pygame.K_RETURN:
                    server_ip = ip_input.strip() if ip_input.strip() else "127.0.0.1"
                    sub_socket, push_socket, req_socket = connect_locally(server_ip)
                    #time.sleep(1)
                    assigned_role, player_id = connect_client_id()
                    #player_num = 2
                    waiting = False
                    text_input_active = False
                elif event.key == pygame.K_BACKSPACE:
                    ip_input = ip_input[:-1]
                else:
                    if len(ip_input) < 30 and (event.unicode.isdigit() or event.unicode == '.' or event.unicode == ':'):
                        ip_input += event.unicode
        

        # Draw menu every frame (styled to match pause UI)
        screen.fill((0,0,0))
        panel = pygame.Rect(WIDTH//2 - 250, HEIGHT//2 - 150, 500, 300)
        ui_draw_panel(screen, panel)

        title_font = pygame.font.SysFont(None, 64)
        title = title_font.render("Pong Game", True, (240,240,240))
        screen.blit(title, title.get_rect(midtop=(panel.centerx, panel.y + 20)))

        if text_input_active:
            # input box styled
            pygame.draw.rect(screen, (40,40,40), input_box_rect, border_radius=8)
            pygame.draw.rect(screen, (200,200,200), input_box_rect, 2, border_radius=8)
            ip_surf = font.render(ip_input, True, input_text_color)
            screen.blit(ip_surf, (input_box_rect.x + 8, input_box_rect.y + 5))
        else:
            ui_draw_btn(screen, buttonConnect_rect, "Connect to Server", 30, hot=False)
            ui_draw_btn(screen, buttonServer_rect,  "Start Local Server", 30, hot=False)

        pygame.display.flip()

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
            print
            if msg["type"] == "state":
                remote_state = msg
    except zmq.Again:
        pass

    # --- handle input
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            running = False

            # Open pause menu with ESC or P (only after the waiting screen)
        if not waiting and event.type == pygame.KEYDOWN and event.key in (pygame.K_ESCAPE, pygame.K_p):
            result = show_pause_menu(screen, push_socket, client_id)
            if result == "quit":
                cleanup_and_exit()
            continue  # skip the rest of this frame while paused

            # Only handle paddle input if we're a player
        if assigned_role == "player":
            keys = pygame.key.get_pressed()
            if keys:
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
    
        # Only handle input if we're a player
        if assigned_role == "player":
            keys = pygame.key.get_pressed()
            # Use different controls for Player 1 vs Player 2
            if keys:
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
                print(f"Sent input: up={up}, down={down}, client_id={client_id}, player_id={player_id}")

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
    