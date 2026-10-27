package main

import (
	"errors"
	"fmt"
	"lesson5/student/practice"
)

func main() {
	// TODO: домашняя работа. Покажите успешную рассылку, пустой текст,
	fmt.Println("=== 1. Успешная отправка ===")

	email := &practice.EmailNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "email"},
		Address:      "student@example.com",
	}

	push := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "device-123",
	}

	err := practice.Dispatch(
		[]practice.Notifier{email, push},
		"Заказ готов",
	)

	if err == nil {
		fmt.Println("Уведомления отправлены успешно")
	} else {
		fmt.Println("Ошибка:", err)
	}

	fmt.Println("Email:", email.Messages)
	fmt.Println("Push:", push.Messages)

	fmt.Println("\n=== 2. Пустое сообщение ===")

	err = practice.Dispatch(
		[]practice.Notifier{email, push},
		"",
	)

	if errors.Is(err, practice.ErrEmptyMessage) {
		fmt.Println("Обнаружено пустое сообщение:", err)
	} else if err != nil {
		fmt.Println("Другая ошибка:", err)
	}

	fmt.Println("Email:", email.Messages)
	fmt.Println("Push:", push.Messages)

	fmt.Println("\n=== 3. Нет DeviceID ===")

	badPush := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
	}

	err = practice.Dispatch(
		[]practice.Notifier{badPush},
		"Привет",
	)

	var sendErr *practice.SendError
	if errors.As(err, &sendErr) {
		fmt.Println("Ошибка настройки")
		fmt.Println("Канал:", sendErr.Channel)
		fmt.Println("Причина:", sendErr.Reason)
	} else if err != nil {
		fmt.Println("Другая ошибка:", err)
	}

	fmt.Println("История:", badPush.Messages)

	fmt.Println("\n=== 4. Паника одного канала ===")

	broken := &practice.BrokenNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "broken"},
	}

	goodPush := &practice.PushNotifier{
		BaseNotifier: practice.BaseNotifier{Channel: "push"},
		DeviceID:     "device-456",
	}

	failures := practice.SafeDispatch(
		[]practice.Notifier{broken, goodPush},
		"Заказ готов",
	)

	for _, failure := range failures {
		if errors.Is(failure, practice.ErrNotifierPanic) {
			fmt.Println("Перехвачена паника:", failure)
		} else {
			fmt.Println("Ошибка отправки:", failure)
		}
	}

	fmt.Println("Push после паники:", goodPush.Messages)
}
