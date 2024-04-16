package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"mxshop_api/user-web/forms"
	"net/http"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"mxshop_api/user-web/global"
)

func GenerateSmsCode() string {
	rand.Seed(time.Now().UnixNano())
	var code strings.Builder
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&code, "%d", rand.Intn(10))
	}
	return code.String()
}

func CreateClient() (_result *dysmsapi20170525.Client, _err error) {
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考。
	// 建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html。
	config := &openapi.Config{
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: tea.String(global.ServerConfig.AliSmsInfo.ApiKey),
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: tea.String(global.ServerConfig.AliSmsInfo.ApiSecrect),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

func SendSms(ctx *gin.Context) {
	// 表单验证
	smsForm := forms.SendSmsForm{}
	if err := ctx.ShouldBind(&smsForm); err != nil {
		HandleValidatorError(ctx, err)
		return
	}

	client, _err := CreateClient()
	if _err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "验证码错误",
		})
		return
	}
	templateCode := "SMS_267940363"
	code := GenerateSmsCode()
	templateParam := fmt.Sprintf(`{"code":"%s"}`, code)
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(smsForm.Mobile),
		SignName:      tea.String("谷粒商城"),
		TemplateCode:  &templateCode,
		TemplateParam: &templateParam,
	}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		resp, _err := client.SendSmsWithOptions(sendSmsRequest, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		d, _ := json.Marshal(resp)
		fmt.Println(string(d))
		// 将验证码和手机号保存下来
		rdb := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
		})
		rdb.Set(context.Background(), smsForm.Mobile, code, time.Duration(global.ServerConfig.AliSmsInfo.Expire)*time.Second)
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "发送成功",
		})
		return nil
	}()

	if tryErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "发送验证码异常",
		})
	}
}
