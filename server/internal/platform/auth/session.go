package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	sessionmodel "shijibu/internal/model/redis"
)

var (
	ErrUnauthorized     = errors.New("invalid or expired session")
	ErrStoreUnavailable = errors.New("session store unavailable")
)

type SessionStore interface {
	CreateSession(context.Context, string, sessionmodel.Session, time.Duration) error
	MatchSession(context.Context, string, string, string) (bool, error)
	RotateSession(context.Context, string, string, sessionmodel.Session, time.Duration) error
}

type SessionManager struct {
	secret []byte
	ttl    time.Duration
	store  SessionStore
}

type TokenPair struct {
	Token            string `json:"token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

type sessionClaims struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"sid"`
	Kind      string `json:"kind"`
	jwt.RegisteredClaims
}

func NewSessionManager(secret string, ttl time.Duration, store SessionStore) *SessionManager {
	if len(secret) < 32 || ttl <= 0 || store == nil {
		panic("invalid session manager configuration")
	}
	return &SessionManager{secret: []byte(secret), ttl: ttl, store: store}
}

// tokenHash 计算 Token 的 SHA-256 十六进制摘要，供 Redis 保存和后续匹配。
// 摘要不可用于还原凭据，避免 Redis 中直接存放可发起请求的 Token 原文。
func tokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// pair 为同一用户和会话生成访问、刷新两种 JWT，并返回用于 Redis 的摘要数据。
// 本函数只生成凭据，不写 Redis；两个 JWT 使用相同有效期，以 kind 区分用途。
func (m *SessionManager) pair(userID uint, id string) (*TokenPair, sessionmodel.Session, error) {
	now := time.Now()
	// 每个 JWT 使用独立的 jti，避免同一秒内重复签发相同凭据；sid 则关联共同会话。
	sign := func(kind string) (string, error) {
		claims := sessionClaims{UserID: userID, SessionID: id, Kind: kind, RegisteredClaims: jwt.RegisteredClaims{
			ID: uuid.NewString(), Subject: strconv.FormatUint(uint64(userID), 10), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		}}
		return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	}
	access, err := sign("access") // 签发供业务接口鉴权的访问 Token。
	if err != nil {
		return nil, sessionmodel.Session{}, err
	}
	refresh, err := sign("refresh") // 签发仅用于刷新会话的刷新 Token。
	if err != nil {
		return nil, sessionmodel.Session{}, err
	}
	return &TokenPair{access, refresh, int64(m.ttl / time.Second), int64(m.ttl / time.Second)}, sessionmodel.Session{UserID: userID, AccessHash: tokenHash(access), RefreshHash: tokenHash(refresh)}, nil // tokenHash 生成摘要，Redis 不保存 Token 原文。
}

// Issue 创建一次独立登录会话，只有凭据摘要成功写入 Redis 后才返回双 Token。
// 不同登录拥有不同会话 ID；Redis 不可用时返回 ErrStoreUnavailable，禁止退回纯 JWT 登录。
func (m *SessionManager) Issue(ctx context.Context, userID uint) (*TokenPair, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}
	id := uuid.NewString()
	pair, session, err := m.pair(userID, id) // 生成同一会话的双 Token 及其存储摘要。
	if err != nil {
		return nil, err
	}
	if err := m.store.CreateSession(ctx, id, session, m.ttl); err != nil { // 将摘要写入 Redis，并设置会话有效期。
		return nil, ErrStoreUnavailable
	}
	return pair, nil
}

func (m *SessionManager) claims(value, kind string) (*sessionClaims, error) {
	claims := &sessionClaims{}
	token, err := jwt.ParseWithClaims(value, claims, func(*jwt.Token) (any, error) { return m.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid || claims.UserID == 0 || claims.SessionID == "" || claims.Kind != kind {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

func (m *SessionManager) Authenticate(ctx context.Context, value string) (uint, error) {
	return m.validate(ctx, value, "access")
}

func (m *SessionManager) ValidateRefresh(ctx context.Context, value string) (uint, error) {
	return m.validate(ctx, value, "refresh")
}

func (m *SessionManager) validate(ctx context.Context, value, kind string) (uint, error) {
	claims, err := m.claims(value, kind)
	if err != nil {
		return 0, err
	}
	match, err := m.store.MatchSession(ctx, claims.SessionID, kind+"_hash", tokenHash(value))
	if err != nil {
		return 0, ErrStoreUnavailable
	}
	if !match {
		return 0, ErrUnauthorized
	}
	return claims.UserID, nil
}

func (m *SessionManager) Refresh(ctx context.Context, value string) (*TokenPair, error) {
	claims, err := m.claims(value, "refresh")
	if err != nil {
		return nil, err
	}
	pair, session, err := m.pair(claims.UserID, claims.SessionID)
	if err != nil {
		return nil, err
	}
	err = m.store.RotateSession(ctx, claims.SessionID, tokenHash(value), session, m.ttl)
	if errors.Is(err, sessionmodel.ErrSessionMissing) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, ErrStoreUnavailable
	}
	return pair, nil
}
