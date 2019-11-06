package aliacm

import (
	"fmt"
	"reflect"
)

type ConcreteObserver interface {
	Modify(unit Unit, config Config)
}

type observer struct {
	observerMap map[string]map[string][]ConcreteObserver
}

func Observer() *observer {
	var obs observer
	obs.observerMap = make(map[string]map[string][]ConcreteObserver)
	return &obs
}

func (o *observer) Modify(unit Unit, config Config) {
	for _, obs := range o.observerMap[unit.Group][unit.DataID] {
		obs.Modify(unit, config)
	}
}

func (o *observer) Register(groupID, dataID string, co ConcreteObserver) error {
	if reflect.TypeOf(co).Kind() != reflect.Ptr {
		return fmt.Errorf("ConcreteObserver type error")
	}
	if _, ok := o.observerMap[groupID]; !ok {
		o.observerMap[groupID] = make(map[string][]ConcreteObserver)
	}
	o.observerMap[groupID][dataID] = append(o.observerMap[groupID][dataID], co)
	return nil
}
