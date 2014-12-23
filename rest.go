package prostor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	DEFAULT_URL      = "http://api.prostor-sms.ru/messages/v2"
	DEFAULT_USERNAME = ""
	DEFAULT_PASSWORD = ""
)

var (
	Url      string = DEFAULT_URL
	Username string = DEFAULT_USERNAME
	Password string = DEFAULT_PASSWORD

	err error
)

type Client string

func urlencode(data map[string]interface{}) string {
	if len(data) <= 0 {
		return ""
	}

	var buf bytes.Buffer
	for k, v := range data {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(fmt.Sprintf("%v", v)))
		buf.WriteByte('&')
	}

	mData := buf.String()

	return mData[0 : len(mData)-1]
}

// Получения состояния счета (RUB;540.15;0.0:: 1 значение – тип баланса, 2 значение – баланс, 3 значение – кредит)
func (c *Client) GetBalance() (string, error) {
	var params = map[string]interface{}{}

	balance, err := request("GET", "/balance/", params)
	if err != nil {
		return "", err
	}

	return balance, nil
}

// Отправка сообщения
// Статусы:
// accepted - Сообщение принято сервисом
// invalid mobile phone - Неверно задан номер тефона (формат +71234567890)
// text is empty - Отсутствует текст
// sender address invalid - Неверная (незарегистрированная) подпись отправителя
// wapurl invalid - Неправильный формат wap-push ссылки
// invalid schedule time format - Неверный формат даты отложенной отправки сообщения (в UTC (2008-07-12T14:30:01Z))
// invalid status queue name - Неверное название очереди статусов сообщений
func (c *Client) SendMessage(src_address, dst_address, data string, send_date_utc string) (string, error) {
	var params = map[string]interface{}{
		"sender":       src_address,
		"phone":        dst_address,
		"text":         data,
		"scheduleTime": send_date_utc,
	}

	req, err := request("GET", "/send/?", params)
	if err != nil {
		return "", err
	}
	res := strings.Split(req, ";")
	if res[0] != "accepted" {
		switch res[1] {
		case "invalid mobile phone":
			return "", fmt.Errorf("Неверно задан номер тефона (формат +71234567890)")
		case "text is empty":
			return "", fmt.Errorf("Отсутствует текст")
		case "absent required param: text":
			return "", fmt.Errorf("Отсутствует текст")
		case "sender address invalid":
			return "", fmt.Errorf("Неверная (незарегистрированная) подпись отправителя")
		case "wapurl invalid":
			return "", fmt.Errorf("Неправильный формат wap-push ссылки")
		case "invalid schedule time format":
			return "", fmt.Errorf("Неверный формат даты отложенной отправки сообщения (в UTC (2008-07-12T14:30:01Z))")
		case "invalid status queue name":
			return "", fmt.Errorf("Неверное название очереди статусов сообщений")
		default:
			return "", fmt.Errorf("Неизвестная ошибка")
		}
	}

	return res[1], nil
}

// Проверка статуса сообщения
// Статусы:
// queued - Сообщение находится в очереди
// delivered - Сообщение доставлено
// delivery error - Ошибка доставки SMS (абонент в течение времени доставки находился вне зоны действия сети или номер абонента заблокирован)
// smsc submit - Сообщение доставлено в SMSC
// smsc reject - Сообщение отвергнуто SMSC (номер заблокирован или не существует)
// incorrect id - Неверный идентификатор сообщения
func (c *Client) GetMessageState(message_id string) (string, error) {
	var params = map[string]interface{}{
		"id": message_id,
	}

	req, err := request("GET", "/status/?", params)
	if err != nil {
		return "", err
	}

	res := strings.Split(req, ";")
	if res[0] != "" {
		switch res[1] {
		case "queued":
			return "Сообщение находится в очереди", nil
		case "delivered":
			return "Сообщение доставлено", nil
		case "delivery error":
			return "Ошибка доставки SMS", nil
		case "smsc submit":
			return "Сообщение доставлено в SMSC", nil
		case "smsc reject":
			return "Сообщение отвергнуто SMSC (номер заблокирован или не существует)", nil
		case "incorrect id":
			return "Неверный идентификатор сообщения", nil
		default:
			return "Статус неизвестен", nil
		}
	}

	return "", nil
}

// Список доступных подписей отправителя
func (c *Client) GetSenders() (string, error) {
	var params = map[string]interface{}{}

	balance, err := request("GET", "/senders/", params)
	if err != nil {
		return "", err
	}

	return balance, nil
}

// Проверерка активной версии API
func (c *Client) GetApiVersion() (string, error) {
	var params = map[string]interface{}{}

	api, err := request("GET", "/version/", params)
	if err != nil {
		return "", err
	}

	return api, nil
}

func request(method, path string, params map[string]interface{}) (string, error) {
RequestStart:
	dst := fmt.Sprintf("%s%s%s", Url, path, urlencode(params))
	req, err := http.NewRequest(method, dst, bytes.NewBuffer([]byte("")))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(Username, Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode == 401 {
		response.Body.Close()

		goto RequestStart
	}

	body, _ := ioutil.ReadAll(response.Body)

	return string(body), nil
}
