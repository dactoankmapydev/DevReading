package handler

import (
	"backend-viblo-trending/custom_error"
	"backend-viblo-trending/model"
	"backend-viblo-trending/model/req"
	"backend-viblo-trending/repository"
	"backend-viblo-trending/security"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

type smtpServer struct {
	host string
	port string
}

// Address URI to smtp server.
func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

type UserHandler struct {
	UserRepo repository.UserRepo
	AuthRepo repository.AuthRepo
}

func (u *UserHandler) SignUp(c echo.Context) error {
	request := req.ReqtSignUp{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	hash := security.HashAndSalt([]byte(request.Password))

	userID, err := uuid.NewUUID()
	if err != nil {
		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	user := model.User{
		UserID:   userID.String(),
		FullName: request.FullName,
		Email:    request.Email,
		Password: hash,
		Verify:   false,
	}

	user, err = u.UserRepo.SaveUser(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusConflict, model.Response{
			StatusCode: http.StatusConflict,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	// verify email
	token := security.CreateTokenHash(user.Email)

	// save token to redis
	saveErr := u.AuthRepo.CreateAuthMail(token, user.UserID)
	if saveErr != nil {
		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    saveErr.Error(),
			Data:       nil,
		})
	}

	link := "http://127.0.0.1:4000" + "/user/verify?token=" + token

	from := os.Getenv("FROM")
	password := os.Getenv("PASSWORD")
	to := []string{user.Email}

	smtpsv := smtpServer{
		host: os.Getenv("SMTP_HOST"),
		port: os.Getenv("SMTP_PORT"),
	}

	subject := "Xác thực tài khoản\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := "Để xác thực tài khoản nhấp vào liên kết <a href='" + link + "'>ở đây</a>."
	message := []byte("Subject:" + subject + mime + "\r\n" + body)

	auth := smtp.PlainAuth("", from, password, smtpsv.host)
	errSendMail := smtp.SendMail(smtpsv.Address(), auth, from, to, message)
	if errSendMail != nil {
		return errSendMail
	}

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Tin nhắn xác thực tài khoản được gửi đến email được cung cấp. Vui lòng kiểm tra thư mục thư rác",
	})
}

func (u *UserHandler) ForgotPassword(c echo.Context) error {
	request := req.ReqEmail{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	user, err := u.UserRepo.CheckEmail(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	token := security.CreateTokenHash(user.Email)

	// save token to redis
	saveErr := u.AuthRepo.CreateAuthMail(token, user.UserID)
	if saveErr != nil {
		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    saveErr.Error(),
			Data:       nil,
		})
	}

	insertErr := u.AuthRepo.InsertAuthMail(token)
	if insertErr != nil {
		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    insertErr.Error(),
			Data:       nil,
		})
	}

	link := "http://127.0.0.1:4000" + "/user/password/reset?token=" + token

	from := os.Getenv("FROM")
	password := os.Getenv("PASSWORD")
	to := []string{user.Email}

	smtpsv := smtpServer{
		host: os.Getenv("SMTP_HOST"),
		port: os.Getenv("SMTP_PORT"),
	}

	subject := "Đặt lại mật khẩu\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := "Để đặt lại mật khẩu nhấp vào liên kết <a href='" + link + "'>ở đây</a>."
	message := []byte("Subject:" + subject + mime + "\r\n" + body)

	auth := smtp.PlainAuth("", from, password, smtpsv.host)
	errSendMail := smtp.SendMail(smtpsv.Address(), auth, from, to, message)
	if errSendMail != nil {
		return errSendMail
	}

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Tin nhắn được gửi đến email được cung cấp. Vui lòng kiểm tra thư mục thư rác",
	})
}

func (u *UserHandler) VerifyAccount(c echo.Context) error {
	request := req.PasswordSubmit{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	token := security.ExtractTokenMail(c.Request())

	userID, err := u.AuthRepo.FetchAuthMail(token)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Truy cập không được phép",
			Data:       nil,
		})
	}

	user, err := u.UserRepo.SelectUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if request.Password != request.Confirm {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Xác nhận mật khẩu không khớp",
			Data:       nil,
		})
	}

	// check password
	isTheSame := security.ComparePasswords(user.Password, []byte(request.Password))
	if !isTheSame {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Mật khẩu không đúng",
			Data:       nil,
		})
	}

	user = model.User{
		UserID: userID,
		Verify: true,
	}

	user, err = u.UserRepo.UpdateVerify(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	deleteAtErr := u.AuthRepo.DeleteTokenMail(token)
	if deleteAtErr != nil {
		log.Println(deleteAtErr)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    deleteAtErr.Error(),
			Data:       nil,
		})
	}

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Xác thực tài khoản thành công",
	})

}

func (u *UserHandler) ResetPassword(c echo.Context) error {
	request := req.PasswordSubmit{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	token := security.ExtractTokenMail(c.Request())

	userID, err := u.AuthRepo.FetchAuthMail(token)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Truy cập không được phép, cần gửi lại email",
			Data:       nil,
		})
	}

	if request.Password != request.Confirm {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Xác nhận mật khẩu không khớp",
			Data:       nil,
		})
	}

	hash := security.HashAndSalt([]byte(request.Password))

	user := model.User{
		UserID:   userID,
		Password: hash,
	}

	user, err = u.UserRepo.UpdatePassword(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	deleteAtErr := u.AuthRepo.DeleteTokenMail(token)
	if deleteAtErr != nil {
		log.Println(deleteAtErr)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    deleteAtErr.Error(),
			Data:       nil,
		})
	}

	return c.JSON(http.StatusCreated, model.Response{
		StatusCode: http.StatusCreated,
		Message:    "Cập nhật mật khẩu thành công",
	})
}

func (u *UserHandler) SignIn(c echo.Context) error {
	request := req.ReqSignIn{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	user, err := u.UserRepo.CheckSignIn(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if user.Verify != true {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Tài khoản chưa được xác thực",
			Data:       nil,
		})
	}

	// check password
	isTheSame := security.ComparePasswords(user.Password, []byte(request.Password))
	if !isTheSame {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Mật khẩu không đúng",
			Data:       nil,
		})
	}

	// create token
	token, err := security.CreateToken(user.UserID)
	if err != nil {
		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	saveErr := u.AuthRepo.CreateAuth(user.UserID, token)
	if saveErr != nil {
		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    saveErr.Error(),
			Data:       nil,
		})
	}

	user.AccessToken = token.AccessToken
	user.RefreshToken = token.RefreshToken

	// create cookie for client(browser)
	atCookie := &http.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		HttpOnly: true,
		SameSite: 2,
		Expires:  time.Now().Add(time.Minute * 15),
	}
	c.SetCookie(atCookie)

	rtCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token.RefreshToken,
		SameSite: 2,
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 24),
	}
	c.SetCookie(rtCookie)

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Đăng nhập thành công",
		Data:       user,
	})
}

func (u *UserHandler) Profile(c echo.Context) error {
	tokenAuth, err := security.ExtractAccessTokenMetadata(c.Request())
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	userID, err := u.AuthRepo.FetchAuth(tokenAuth.AccessUUID)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Truy cập không được phép",
			Data:       nil,
		})
	}

	user, err := u.UserRepo.SelectUserByID(c.Request().Context(), userID)
	if err != nil {
		log.Println(err)
		if err == custom_error.UserNotFound {
			return c.JSON(http.StatusNotFound, model.Response{
				StatusCode: http.StatusNotFound,
				Message:    err.Error(),
				Data:       nil,
			})
		}

		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Xử lý thành công",
		Data:       user,
	})
}

func (u *UserHandler) UpdateProfile(c echo.Context) error {
	request := req.ReqUpdateUser{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	tokenAuth, err := security.ExtractAccessTokenMetadata(c.Request())
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	userID, err := u.AuthRepo.FetchAuth(tokenAuth.AccessUUID)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Truy cập không được phép",
			Data:       nil,
		})
	}

	if request.Password != request.Confirm {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Xác nhận mật khẩu không khớp",
			Data:       nil,
		})
	}

	hash := security.HashAndSalt([]byte(request.Password))

	if request.FullName == "" {
		if len(request.Password) < 8 {
			return c.JSON(http.StatusBadRequest, model.Response{
				StatusCode: http.StatusBadRequest,
				Message:    "Mật khẩu tối thiểu 8 ký tự",
				Data:       nil,
			})
		}
		user := model.User{
			UserID:   userID,
			Password: hash,
		}

		user, err = u.UserRepo.UpdateUser(c.Request().Context(), user)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, model.Response{
				StatusCode: http.StatusUnprocessableEntity,
				Message:    err.Error(),
				Data:       nil,
			})
		}
	}

	if request.Password == "" {
		user := model.User{
			UserID:   userID,
			FullName: request.FullName,
		}

		user, err = u.UserRepo.UpdateUser(c.Request().Context(), user)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, model.Response{
				StatusCode: http.StatusUnprocessableEntity,
				Message:    err.Error(),
				Data:       nil,
			})
		}
	}

	user := model.User{
		UserID:   userID,
		FullName: request.FullName,
		Password: hash,
	}

	user, err = u.UserRepo.UpdateUser(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	return c.JSON(http.StatusCreated, model.Response{
		StatusCode: http.StatusCreated,
		Message:    "Cập nhật thông tin thành công",
	})
}

func (u *UserHandler) SignOut(c echo.Context) error {
	extractAt, err := security.ExtractAccessTokenMetadata(c.Request())
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	deleteAtErr := u.AuthRepo.DeleteAccessToken(extractAt.AccessUUID)
	if deleteAtErr != nil {
		log.Println(deleteAtErr)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    deleteAtErr.Error(),
			Data:       nil,
		})
	}

	extractRt, err := security.ExtractRefreshTokenMetadata(c.Request())
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	deleteRtErr := u.AuthRepo.DeleteRefreshToken(extractRt.RefreshUUID)
	if deleteRtErr != nil {
		log.Println(deleteRtErr)
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    deleteRtErr.Error(),
			Data:       nil,
		})
	}

	atCookie := &http.Cookie{
		Name:   "access_token",
		MaxAge: -1,
	}
	c.SetCookie(atCookie)

	rtCookie := &http.Cookie{
		Name:   "refresh_token",
		MaxAge: -1,
	}
	c.SetCookie(rtCookie)

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Đăng xuất thành công",
	})
}

func (u *UserHandler) Refresh(c echo.Context) error {
	_, err := c.Cookie("access_token")
	if err != nil {
		extractRt, err := security.ExtractRefreshTokenMetadata(c.Request())
		if err != nil {
			return c.JSON(http.StatusUnauthorized, model.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    "Bạn cần phải đăng nhập",
			})
		}

		deleteErr := u.AuthRepo.DeleteRefreshToken(extractRt.RefreshUUID)
		if deleteErr != nil {
			log.Println(deleteErr)
			return c.JSON(http.StatusUnauthorized, model.Response{
				StatusCode: http.StatusUnauthorized,
				Message:    deleteErr.Error(),
				Data:       nil,
			})
		}

		token, createErr := security.CreateToken(extractRt.UserID)
		if createErr != nil {
			return c.JSON(http.StatusForbidden, model.Response{
				StatusCode: http.StatusForbidden,
				Message:    createErr.Error(),
				Data:       nil,
			})
		}

		saveErr := u.AuthRepo.CreateAuth(extractRt.UserID, token)
		if saveErr != nil {
			return c.JSON(http.StatusForbidden, model.Response{
				StatusCode: http.StatusForbidden,
				Message:    saveErr.Error(),
				Data:       nil,
			})
		}

		tokens := map[string]string{
			"access_token":  token.AccessToken,
			"refresh_token": token.RefreshToken,
		}

		atCookie := &http.Cookie{
			Name:     "access_token",
			Value:    token.AccessToken,
			HttpOnly: true,
			SameSite: 2,
			Expires:  time.Now().Add(time.Minute * 15),
		}
		c.SetCookie(atCookie)

		rtCookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    token.RefreshToken,
			SameSite: 2,
			HttpOnly: true,
			Expires:  time.Now().Add(time.Hour * 24),
		}
		c.SetCookie(rtCookie)

		return c.JSON(http.StatusCreated, model.Response{
			StatusCode: http.StatusCreated,
			Message:    "Xử lý thành công",
			Data:       tokens,
		})
	}

	return c.JSON(http.StatusUnauthorized, model.Response{
		StatusCode: http.StatusUnauthorized,
		Message:    "Access token chưa hết hạn",
	})
}
