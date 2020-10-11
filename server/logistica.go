package main
import (
	"fmt"
	"time"
	"reflect"
)
var id=0
type orden struct {
	timestamp time.Time
	idPaquete string
	tipo string
	nombre string
	valor int
	origen string
	destino string
	seguimiento string
}
func newOrden(tipo string,nombre string,valor int,origen string, destino string) *orden{
	ordenNueva := orden{nombre: nombre,tipo: tipo,valor: valor ,origen: origen, destino: destino}
	ordenNueva.timestamp= time.Now()
	id:=id+1
	ordenNueva.idPaquete = fmt.Sprint(id)
	ordenNueva.seguimiento=fmt.Sprint(id*407)
	return &ordenNueva
}
type paquete struct {
	idPaquete string
	tipo string
	valor int
	seguimiento string
	intentos int
	estado string
}

func newPaquete(idPaquete string, tipo string, valor int, seguimiento string) *paquete{
	paqueteNuevo := paquete{idPaquete: idPaquete, tipo: tipo,valor: valor,seguimiento: seguimiento}
	paqueteNuevo.intentos= 0;
	paqueteNuevo.estado="en bodega"
	return &paqueteNuevo
}
func main(){
	t:= time.Now()
	fmt.Println(reflect.TypeOf(t))
}
