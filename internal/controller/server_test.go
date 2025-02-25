package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DmitryM7/yapr56.git/internal/conf"
	"github.com/DmitryM7/yapr56.git/internal/logger"
	mock_controller "github.com/DmitryM7/yapr56.git/internal/mocks"
	"github.com/DmitryM7/yapr56.git/internal/models"
	"github.com/DmitryM7/yapr56.git/internal/sec"
	"github.com/DmitryM7/yapr56.git/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	gconf   conf.Config
	glogger logger.Lg
	gjwt    sec.JwtProvider
)

func TestMain(m *testing.M) {
	gconf = conf.NewConf()
	glogger = logger.NewLg()
	gjwt = sec.NewJwtProvider(gconf.SecretKeyTime, gconf.SecretKey)
	os.Exit(m.Run())
}

func TestSrv_actUserRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageservice := mock_controller.NewMockIStorage(ctrl)

	storageservice.EXPECT().CreatePeson(gomock.Any(), models.Person{
		Login: "dmaslov",
		Pass:  "!QAZ2wsx",
	}).MaxTimes(1).Return(models.Person{
		Login: "dmaslov",
		Pass:  "!QAZ2wsx",
	}, nil)

	storageservice.EXPECT().CreatePeson(gomock.Any(), models.Person{
		Login: "dmaslov",
		Pass:  "!QAZ2wsx",
	}).MaxTimes(1).Return(models.Person{
		Login: "dmaslov",
		Pass:  "!QAZ2wsx",
	}, service.ErrUserExists)

	serv, err := NewServer(glogger, storageservice, gjwt)
	if err != nil {
		t.Fatalf("TEST ERROR. CAN'T CREATE SERVER: [%v]", err)
	}

	type args struct {
		w      *httptest.ResponseRecorder
		r      *http.Request
		person UserRegisterRequest
		method string
	}

	type want struct {
		StatusCode int
	}

	tests := []struct {
		name string
		s    *Srv
		args args
		want want
	}{
		{
			name: "Person can register",
			s:    serv,
			args: args{
				method: http.MethodPost,
				person: UserRegisterRequest{
					Login:    "dmaslov",
					Password: "!QAZ2wsx",
				},
			},
			want: want{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "Person can't create. User exists.",
			s:    serv,
			args: args{
				method: http.MethodPost,
				person: UserRegisterRequest{
					Login:    "dmaslov",
					Password: "!QAZ2wsx",
				},
			},
			want: want{
				StatusCode: http.StatusConflict,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.args.person)
			if err != nil {
				t.Errorf("ERROR IN TEST. INVALID args -> person.")
				return
			}
			tt.args.r = httptest.NewRequest(tt.args.method, "/api/user/register", strings.NewReader(string(body)))
			tt.args.w = httptest.NewRecorder()
			tt.s.actUserRegister(tt.args.w, tt.args.r)

			res := tt.args.w.Result()

			assert.Equal(t, tt.want.StatusCode, res.StatusCode)

		})
	}
}

func TestSrv_actUserLogin(t *testing.T) {

	type args struct {
		w      *httptest.ResponseRecorder
		r      *http.Request
		person UserAuthRequest
		method string
	}

	type want struct {
		StatusCode int
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageservice := mock_controller.NewMockIStorage(ctrl)

	storageservice.EXPECT().GetPesonByCredential(gomock.Any(), "dmaslov", "!QAZ2wsx").MaxTimes(1).Return(models.Person{ID: 1}, nil)
	storageservice.EXPECT().GetPesonByCredential(gomock.Any(), "dmaslov23", "!QAZ2wsx").MaxTimes(1).Return(models.Person{}, service.ErrUserCredentialInvalid)

	serv, err := NewServer(glogger, storageservice, gjwt)

	if err != nil {
		t.Fatalf("TEST ERROR. CAN'T CREATE SERVER: [%v]", err)
	}

	tests := []struct {
		name string
		s    *Srv
		args args
		want want
	}{
		{
			name: "Person can login",
			s:    serv,
			args: args{
				method: http.MethodPost,
				person: UserAuthRequest{
					UserRegisterRequest: UserRegisterRequest{
						Login:    "dmaslov",
						Password: "!QAZ2wsx",
					},
				},
			},
			want: want{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "Person can't login. Invalid credential.",
			s:    serv,
			args: args{
				method: http.MethodPost,
				person: UserAuthRequest{
					UserRegisterRequest: UserRegisterRequest{
						Login:    "dmaslov23",
						Password: "!QAZ2wsx",
					},
				},
			},
			want: want{
				StatusCode: http.StatusUnauthorized,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.args.person)
			if err != nil {
				t.Errorf("ERROR IN TEST. INVALID args -> person.")
				return
			}
			tt.args.r = httptest.NewRequest(tt.args.method, "/api/user/login", strings.NewReader(string(body)))
			tt.args.w = httptest.NewRecorder()
			tt.s.actUserLogin(tt.args.w, tt.args.r)

			res := tt.args.w.Result()

			assert.Equal(t, tt.want.StatusCode, res.StatusCode)
		})
	}
}

func TestSrv_actOrdersUpload(t *testing.T) {
	type args struct {
		w      *httptest.ResponseRecorder
		r      *http.Request
		ExtNum string
		method string
	}

	type want struct {
		StatusCode int
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageservice := mock_controller.NewMockIStorage(ctrl)
	storageservice.EXPECT().GetPersonByID(gomock.Any(), 1).MaxTimes(10).Return(models.Person{ID: 1}, nil)
	storageservice.EXPECT().CreateOrder(gomock.Any(), models.Person{ID: 1}, models.POrder{Extnum: 123}).MaxTimes(1).Return(models.POrder{}, service.ErrNoLuhnNumber)
	storageservice.EXPECT().CreateOrder(gomock.Any(), models.Person{ID: 1}, models.POrder{Extnum: 4991071609492511}).MaxTimes(1).Return(models.POrder{Extnum: 1001005050}, nil)
	storageservice.EXPECT().CreateOrder(gomock.Any(), models.Person{ID: 1}, models.POrder{Extnum: 4991071609492511}).MaxTimes(1).Return(models.POrder{Extnum: 1001005050}, service.ErrDublicateOrder)
	storageservice.EXPECT().CreateOrder(gomock.Any(), models.Person{ID: 1}, models.POrder{Extnum: 4180973527225074}).MaxTimes(1).Return(models.POrder{Extnum: 1001005050}, service.ErrOrderExists)

	serv, err := NewServer(glogger, storageservice, gjwt)

	if err != nil {
		t.Fatalf("TEST ERROR. CAN'T CREATE SERVER: [%v]", err)
	}

	tests := []struct {
		name string
		s    *Srv
		args args
		want want
	}{
		{
			name: "Not Luht number.",
			s:    serv,
			args: args{
				method: http.MethodPost,
				ExtNum: "123",
			},
			want: want{
				StatusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "Upload OK",
			s:    serv,
			args: args{
				method: http.MethodPost,
				ExtNum: "4991071609492511",
			},
			want: want{
				StatusCode: http.StatusAccepted,
			},
		},
		{
			name: "Upload Dublicate",
			s:    serv,
			args: args{
				method: http.MethodPost,
				ExtNum: "4991071609492511",
			},
			want: want{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "Upload Order Exists",
			s:    serv,
			args: args{
				method: http.MethodPost,
				ExtNum: "4180973527225074",
			},
			want: want{
				StatusCode: http.StatusConflict,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.WithValue(context.Background(), contextParam("CurrPersonID"), 1)

			tt.args.r = httptest.NewRequest(tt.args.method, "/api/user/orders", strings.NewReader(tt.args.ExtNum)).WithContext(ctx)
			tt.args.w = httptest.NewRecorder()
			tt.s.actOrdersUpload(tt.args.w, tt.args.r)

			res := tt.args.w.Result()

			assert.Equal(t, tt.want.StatusCode, res.StatusCode)
		})
	}
}
