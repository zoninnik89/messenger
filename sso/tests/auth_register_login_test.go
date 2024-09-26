package tests

import (
	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/zoninnik89/messenger/common/api"
	suite "github.com/zoninnik89/messenger/sso/tests/suite"
	"testing"
	"time"
)

const (
	emptyAdID = 0
	appID     = 1
	appSecret = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	loginTime := time.Now()

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_Login_RepeatedRegister(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegisterLogin_Login_RegisterWithNoEmail(t *testing.T) {
	ctx, st := suite.New(t)

	email := ""
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "email required")
}

func TestRegisterLogin_Login_RegisterWithNoPassword(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := ""

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "password required")
}

func TestRegisterLogin_Login_LoginEmptyApp(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	loginResp, err := st.AuthClient.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    emptyAdID,
	})

	require.Error(t, err)
	assert.Empty(t, loginResp.GetToken())
	assert.ErrorContains(t, err, "app id required")
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
