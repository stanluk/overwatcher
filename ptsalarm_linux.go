//build +linux
package main

type ptsalarm struct{}

func (p *ptsalarm) AlarmNow(msg string) error {
	return nil
}

func init() {
	a := &ptsalarm{}
	InstallAlarmer(a)
}
