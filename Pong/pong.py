import pygame
import sys

# Initialize pygame
pygame.init()

# Screen settings
WIDTH, HEIGHT = 800, 600
WIN = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Pong Game")

# Colors
WHITE = (186, 85, 211)
BLACK = (0, 0, 0)

# Game objects
PADDLE_WIDTH, PADDLE_HEIGHT = 10, 100
BALL_SIZE = 15

# Player paddles
left_paddle = pygame.Rect(50, HEIGHT // 2 - PADDLE_HEIGHT // 2, PADDLE_WIDTH, PADDLE_HEIGHT)
right_paddle = pygame.Rect(WIDTH - 50 - PADDLE_WIDTH, HEIGHT // 2 - PADDLE_HEIGHT // 2, PADDLE_WIDTH, PADDLE_HEIGHT)

# Ball
ball = pygame.Rect(WIDTH // 2, HEIGHT // 2, BALL_SIZE, BALL_SIZE)
ball_vel = [10, 10]

# Speeds
PADDLE_SPEED = 12

# Fonts
font = pygame.font.Font(None, 74)

# Scores
left_score = 0
right_score = 0

# Game loop
clock = pygame.time.Clock()
running = True
while running:
    clock.tick(60)  # 60 FPS
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            running = False

    # Keys pressed
    keys = pygame.key.get_pressed()
    if keys[pygame.K_w] and left_paddle.top > 0:
        left_paddle.y -= PADDLE_SPEED
    if keys[pygame.K_s] and left_paddle.bottom < HEIGHT:
        left_paddle.y += PADDLE_SPEED
    if keys[pygame.K_UP] and right_paddle.top > 0:
        right_paddle.y -= PADDLE_SPEED
    if keys[pygame.K_DOWN] and right_paddle.bottom < HEIGHT:
        right_paddle.y += PADDLE_SPEED

    # Ball movement
    ball.x += ball_vel[0]
    ball.y += ball_vel[1]

    # Collision with top and bottom
    if ball.top <= 0 or ball.bottom >= HEIGHT:
        ball_vel[1] = -ball_vel[1]

    # Collision with paddles
    if ball.colliderect(left_paddle) or ball.colliderect(right_paddle):
        ball_vel[0] = -ball_vel[0]

    # Scoring
    if ball.left <= 0:
        right_score += 1
        ball.center = (WIDTH // 2, HEIGHT // 2)
        ball_vel = [10, 10]

    if ball.right >= WIDTH:
        left_score += 1
        ball.center = (WIDTH // 2, HEIGHT // 2)
        ball_vel = [-10, 10]

    # Drawing
    WIN.fill(BLACK)
    pygame.draw.rect(WIN, WHITE, left_paddle)
    pygame.draw.rect(WIN, WHITE, right_paddle)
    pygame.draw.ellipse(WIN, WHITE, ball)
    pygame.draw.aaline(WIN, WHITE, (WIDTH // 2, 0), (WIDTH // 2, HEIGHT))

    # Display scores
    left_text = font.render(str(left_score), True, WHITE)
    right_text = font.render(str(right_score), True, WHITE)
    WIN.blit(left_text, (WIDTH // 4, 20))
    WIN.blit(right_text, (WIDTH * 3 // 4, 20))

    pygame.display.flip()

pygame.quit()
sys.exit()
