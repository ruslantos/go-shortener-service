package queue

import (
	"context"
	"sync"

	"go.uber.org/zap"

	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

type linksStorage interface {
	DeleteUserURLs(ctx context.Context, ids []string, userID string) error
	DeleteUserURL(ctx context.Context, id string, userID string) error
}

type inChan chan string
type doneChan chan struct{}

type QueueService struct {
	linksStorage linksStorage
	inChan       chan string
	doneChan     chan struct{}
}

func NewQueueService(linksStorage linksStorage, inChan inChan, doneChan doneChan) *QueueService {
	return &QueueService{
		linksStorage: linksStorage,
		inChan:       inChan,
		doneChan:     doneChan,
	}
}

func (q *QueueService) DeleteUserUrls(ctx context.Context, ids []string) error {
	userID := getUserIDFromContext(ctx)
	logger.GetLogger().Info("got deleted user urls", zap.Strings("ids", ids))
	//return q.linksStorage.DeleteUserURLs(ctx, ids, userID)

	// канал с данными
	inputCh := q.generator(q.doneChan, ids)

	err := q.markAsDeleted(ctx, inputCh, userID)
	if err != nil {
		return err
	}

	// получаем слайс каналов
	//channels := fanOut(q.doneChan, inputCh)
	//
	//resultCh := fanIn(q.doneChan, channels)
	//
	//idsToDelete := make([][]string, 0, len(resultCh))
	//for id := range resultCh {
	//	idsToDelete = append(idsToDelete, id)
	//}
	//return q.linksStorage.DeleteUserURLs(ctx, idsToDelete, userID)
	return nil
}

func getUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return ""
	}
	//logger.GetLogger().Info("get userID", zap.String("userID", userID))
	return userID
}

func (q *QueueService) generator(doneCh chan struct{}, input []string) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, data := range input {
			select {
			case <-doneCh:
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}

// fanOut принимает канал данных, порождает 10 горутин
func fanOut(doneCh chan struct{}, inputCh chan string) []chan []string {
	// количество горутин add
	numWorkers := 10
	// каналы, в которые отправляются результаты
	channels := make([]chan []string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		// получаем канал из горутины add
		//addResultCh := inputCh
		// отправляем его в слайс каналов
		//channels[i] = addResultCh
	}

	// возвращаем слайс каналов
	return channels
}

// fanIn объединяет несколько каналов resultChs в один.
func fanIn(doneCh chan struct{}, resultChs ...chan []string) chan []string {
	// конечный выходной канал в который отправляем данные из всех каналов из слайса, назовём его результирующим
	finalCh := make(chan []string)

	// понадобится для ожидания всех горутин
	var wg sync.WaitGroup

	// перебираем все входящие каналы
	for _, ch := range resultChs {
		// в горутину передавать переменную цикла нельзя, поэтому делаем так
		chClosure := ch

		// инкрементируем счётчик горутин, которые нужно подождать
		wg.Add(1)

		go func() {
			// откладываем сообщение о том, что горутина завершилась
			defer wg.Done()

			// получаем данные из канала
			for data := range chClosure {
				select {
				// выходим из горутины, если канал закрылся
				case <-doneCh:
					return
				// если не закрылся, отправляем данные в конечный выходной канал
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		// ждём завершения всех горутин
		wg.Wait()
		// когда все горутины завершились, закрываем результирующий канал
		close(finalCh)
	}()

	// возвращаем результирующий канал
	return finalCh
}

func (q *QueueService) markAsDeleted(ctx context.Context, inputCh chan string, userID string) error {
	var errs error
	var wg sync.WaitGroup

	for URL := range inputCh {
		wg.Add(1)
		go func(URL string) {
			defer wg.Done()
			err := q.linksStorage.DeleteUserURL(ctx, URL, userID)
			if err != nil {
				errs = err
			}
			logger.GetLogger().Info("goroutine delete url", zap.String("url", URL))
		}(URL)
	}
	wg.Wait()
	if errs != nil {
		return errs
	}
	return nil
}
