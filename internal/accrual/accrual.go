package accrual

import (
	"encoding/json"
	"io"
	"kaunnikov/internal/logging"
	"kaunnikov/internal/services"
	"net/http"
	"strconv"
)

type RespApi struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func CheckOrders(accrualUrl string) {
	// Забираем из БД номера нерасчитанных заказов
	//(временно выставляем им промежуточный статус, в итоге возвращаем в Processing, либо проставляем обработанный статус Invalid/Processed)
	numbers, err := services.GetNotProcessedOrderNumbers()
	if err != nil {
		logging.Errorf("Ошибка получения необработанных заказов: %s", err)
		return
	}

	// Если нет необработанных заказов - выходим
	if len(numbers) == 0 {
		return
	}

	ch := generator(numbers)

	for number := range ch {
		// Отправляем запрос в систему расчёта баллов
		resp, err := http.Get(accrualUrl + strconv.FormatInt(number, 10))
		if err != nil {
			logging.Errorf("Ошибка отправки запроса в систему расчёта баллов: %s", err)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		// Заказа нет в системе - убираем в Invalid
		switch resp.StatusCode {
		case http.StatusNoContent:
			if err = services.SetInvalidStatusByNumber(number); err != nil {
				logging.Errorf("Не смогли установить статуст Invalid: %s", err)
			}
		case http.StatusOK:
			var s RespApi
			err = json.Unmarshal(body, &s)
			if err != nil {
				logging.Errorf("cannot decode response: %s", err)
				return
			}

			switch s.Status {
			case "INVALID":
				if err = services.SetInvalidStatusByNumber(number); err != nil {
					logging.Errorf("Не смогли установить статуст Invalid: %s", err)
				}
			case "PROCESSED":
				if err = services.SetProcessedStatusByNumber(number, int(s.Accrual*100)); err != nil {
					logging.Errorf("Не смогли установить статуст Processed: %s", err)
				}
			default:
				// Выставляем обратно статус Processing, чтобы взять в работу снова через N времени
				if err = services.SetProcessingStatusByNumber(number); err != nil {
					logging.Errorf("Не смогли установить статуст Processing: %s", err)
				}
			}
		}
	}
}

func generator(numbers []int64) chan int64 {
	processCh := make(chan int64)
	go func() {
		defer close(processCh)
		for _, n := range numbers {
			processCh <- n
		}
	}()
	return processCh
}