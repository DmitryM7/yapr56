package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestSrv_actUserRegister(t *testing.T) {
	conf := conf.NewConf()
	logger := logger.NewLg()

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

	jwt := sec.NewJwtProvider(conf.SecretKeyTime, conf.SecretKey)
	serv, err := NewServer(logger, storageservice, jwt)
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
