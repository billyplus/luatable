package worker

type Worker struct{
	WorkerPool chan chan Job
	JobChan chan Job
	ErrChan chan error
	quit chan bool
}

func NewWorker(workerpool chan chan Job,errChan chan error) *Worker{
	return &Worker{
		WorkerPool: workerpool,
		JobChan: make(chan Job),
		ErrChan: errChan,
		quit: make(chan bool),
	}
}

func (w *Worker)Run(){
	for{
		// 将自己加入到队列中
		w.WorkerPool <- w.JobChan
		select{
		case job:=<- w.JobChan:
			// 从队列中获取任务
			if err:=job();err!=nil{
				w.ErrChan<-err
			}
		case <-w.quit:
			return
		}
	}
}

func (w *Worker)Stop(){
	w.quit<-true
}