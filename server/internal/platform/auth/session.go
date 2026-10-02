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

func tokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (m *SessionManager) pair(userID uint, id string) (*TokenPair, sessionmodel.Session, error) {
	now := time.Now()
	sign := func(kind string) (string, error) {
		claims := sessionClaims{UserID: userID, SessionID: id, Kind: kind, RegisteredClaims: jwt.RegisteredClaims{
			ID: uuid.NewString(), Subject: strconv.FormatUint(uint64(userID), 10), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		}}
		return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	}
	access, err := sign("access")
	if err != nil {
		return nil, sessionmodel.Session{}, err
	}
	refresh, err := sign("refresh")
	if err != nil {
		return nil, sessionmodel.Session{}, err
	}
	return &TokenPair{access, refresh, int64(m.ttl / time.Second), int64(m.ttl / time.Second)}, sessionmodel.Session{UserID: userID, AccessHash: tokenHash(access), RefreshHash: tokenHash(refresh)}, nil
}

func (m *SessionManager) Issue(ctx context.Context, userID uint) (*TokenPair, error) {
	if userID == 0 {
		return nil, ErrUnauthorized
	}
	id := uuid.NewString()
	pair, session, err := m.pair(userID, id)
	if err != nil {
		return nil, err
	}
	if err := m.store.CreateSession(ctx, id, session, m.ttl); err != nil {
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
