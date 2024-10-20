package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/zoninnik89/messenger/common/api"
	suite "github.com/zoninnik89/messenger/sso/tests/suite"
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
		Login:    email,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &pb.LoginRequest{
		Login:    email,
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
		Login:    email,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &pb.RegisterRequest{
		Login:    email,
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
		Login:    email,
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
		Login:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "password required")
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		pass        string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with empty email",
			email:       "",
			pass:        randomFakePassword(),
			appID:       appID,
			expectedErr: "email required",
		},
		{
			name:        "Login with empty password",
			email:       gofakeit.Email(),
			pass:        "",
			appID:       appID,
			expectedErr: "password required",
		},
		{
			name:        "Login with empty email and password",
			email:       "",
			pass:        "",
			appID:       appID,
			expectedErr: "email required",
		},
		{
			name:        "Login with non-matching email",
			email:       gofakeit.Email(),
			pass:        randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login with non-matching password",
			email:       gofakeit.Email(),
			pass:        randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login with non-matching password",
			email:       gofakeit.Email(),
			pass:        randomFakePassword(),
			appID:       emptyAdID,
			expectedErr: "app id required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := gofakeit.Email()
			pass := randomFakePassword()

			respReg, err := st.AuthClient.Register(ctx, &pb.RegisterRequest{
				Login:    email,
				Password: pass,
			})

			require.NoError(t, err)
			assert.NotEmpty(t, respReg.GetUserId())

			respLogin, err := st.AuthClient.Login(ctx, &pb.LoginRequest{
				Login:    tt.email,
				Password: tt.pass,
				AppId:    tt.appID,
			})

			require.Error(t, err)
			assert.Empty(t, respLogin.GetToken())
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
