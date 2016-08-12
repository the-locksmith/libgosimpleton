package parallel

import(

)

type WorkString struct {
	Work

	Talk	chan string
}

func NewStringFeeder(list []string) (work *WorkString) {
	work = &WorkString{}
	work.Initialise(SuggestNumberOfWorkers(uint(len(list))))
	work.Talk = make(chan string, work.SuggestBufferSize(uint(len(list))))

	work.Feed(func() {
		for _, s := range list {
			work.Talk <-s
		}; close(work.Talk)
	})

	return
}
