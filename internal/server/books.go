package server

import (
	books_servicev1 "books_service/gen/go"
	"books_service/internal/logger"
	"books_service/internal/models"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"strings"
	"time"
)

var secretKey = []byte("VerySecretKey2000")

type Claims struct {
	UserID string //`json:"user_id"`
	//Username string `json:"username"`
	//Role     string `json:"role"`
	jwt.RegisteredClaims
}

type BooksService struct {
	books_servicev1.UnimplementedBooksServiceServer
	db *gorm.DB
}

func RegisterBooksService(gRPC *grpc.Server, db *gorm.DB) {
	books_servicev1.RegisterBooksServiceServer(gRPC, &BooksService{db: db})
}

func (as *BooksService) CreateBook(_ context.Context, req *books_servicev1.AuthRequest) (*books_servicev1.AuthResponse, error) {
	zLog := logger.Get()
	zLog.Debug().Any("req", req).Msg("Register (grpc CreateBook)")
	token, err := createJWT(req.UserUid)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to create JWT")
	}
	uid, err := validJWT(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}
	zLog.Debug().Msgf("token=%v, uid=%v", token, uid)

	book := models.Book{
		BID:       uuid.NewString(),
		Label:     req.Label,
		Author:    req.Author,
		Deleted:   false,
		UserUid:   req.UserUid,
		CreatedAt: time.Now(),
	}
	if err := as.db.Create(book).Error; err != nil {
		// Проверка на конкретную ошибку PostgreSQL
		if strings.Contains(err.Error(), "SQLSTATE 42703") {
			zLog.Error().Err(err).Msg("SQL error: column does not exist")
			return nil, status.Error(codes.Internal, "Database schema mismatch")
		}

		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			zLog.Error().Err(err).Msg("SQL error: duplicate key")
			return nil, status.Error(codes.AlreadyExists, "Book already exists")
		}

		zLog.Error().Err(err).Msg("Failed to create book")
		return nil, status.Error(codes.Internal, "Failed to create book in DB")
	}

	return &books_servicev1.AuthResponse{
		Token:   token,
		Message: "Success created book",
	}, nil
}

func (as *BooksService) GetBook(_ context.Context, req *books_servicev1.AuthRequest) (*books_servicev1.AuthResponse, error) {
	zLog := logger.Get()
	zLog.Debug().Any("req", req).Msg("start getBook service")
	//var book models.Book
	//if err := as.db.Where("bid = ?", id).First(&book).Error; err != nil {
	//	return nil, status.Error(codes.NotFound, "Book not found")
	//}
	return &books_servicev1.AuthResponse{
		Token:   "",
		Message: "Success find the book",
	}, nil
}

func (as *BooksService) DeleteBook(_ context.Context, req *books_servicev1.AuthRequest) (*books_servicev1.AuthResponse, error) {
	zLog := logger.Get()
	zLog.Debug().Any("req", req).Msg("Register (grpc auth_service)")

	//if err := as.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
	//	return nil, status.Error(codes.NotFound, "User not found")
	//}
	//if err := checkPasswordHash(user.Pass, userCreds.Pass); err != nil {
	//	return nil, status.Error(codes.Unauthenticated, "Invalid password")
	//}
	////add JWT
	//token, err := createJWT(user.UID)
	//if err != nil {
	//	return nil, status.Error(codes.Internal, "Failed to create JWT")
	//}
	//zLog.Debug().Any("token", token).Msg("Success login")
	return &books_servicev1.AuthResponse{
		Token:   "token",
		Message: "Success created book",
	}, nil
}
func createJWT(UID string) (string, error) {
	// Данные для токена
	claims := Claims{
		UserID: UID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   UID,
		},
	}
	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Подписываем секретным ключом
	return token.SignedString(secretKey)
}

func validJWT(tokenString string) (string, error) {
	claims := &Claims{}
	// Парсим и проверяем токен
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return "", err
	}
	// Проверяем валидность
	if !token.Valid {
		return "", fmt.Errorf("невалидный токен")
	}
	return claims.UserID, nil
}
