package aliacm

// Hook 提供了长轮询失败发生的回调
type Hook func(unit Unit, err error)
