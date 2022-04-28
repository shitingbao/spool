# spool

simple workers pool
一个简易的协程池实现

## 图形解释

![image](https://github.com/shitingbao/spool/model.jpeg)

## 实现原理及内容

1. 复用协程逻辑  
2. 引入一个调度器概念，将所需要执行的方法或者参数，和可执行协程之间进行调度  
3. 使用 context 树的特点，统一退出（不会产生孤儿协程）  
4. 可自定义接受反馈信息  

## 使用例子(example目录下)  

自定义一个反馈接收
  
```go
  type handle struct {
    Num int
  }
  // 只需要实现 HandleMessage 方法,在协程执行完逻辑后，该方法会接收到反馈结果
  func (h *handle) HandleMessage(m spool.Message) error {
    h.Num++
    log.Println("num:", h.Num, " that your body:", string(m.Body), " and err:", m.Err)
    return nil
  }
```

1.普通 pool（传入不同的方法逻辑）

```go
  func poolLoad() {
    h := &handle{}
    option := spool.WithHandlePoolMessage(h)
    p := spool.NewPool(5, option)// 将自定义的结构传入即可
    for i := 0; i < 10; i++ {
      p.Submit(func() (spool.WorkResult, error) {
        return "spool here", nil
      })
    }
    time.Sleep(time.Second * 2)
    p.Release()
    time.Sleep(time.Second * 3)
  }

```

2.固定方法的(PoolWithFunc)  

```go

  func poolWithFuncLoad() {
    h := &handle{}
    option := spool.WithHandlePoolMessage(h)
    p := spool.NewPoolWithFunc(3, func(param interface{}) (spool.WorkResult, error) {
      return "spool", errors.New("this is a error;because err not nil,body is nil")
    }, option)
    for i := 0; i < 10; i++ {
      p.Submit(i)
    }
    time.Sleep(time.Second * 2)
    p.Release()
    time.Sleep(time.Second * 3)
  }

```

## 参考资料  

1. MPG 调度器概念  
2. nsq 消息接收  github.com/nsqio/go-nsq
3. ant 池的定义  github.com/panjf2000/ants  
