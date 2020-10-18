package main
import (
	"time"
	"math/rand"
	"log"
	"net"
	"context"
	"google.golang.org/grpc"
	pb "../proto"
)
const (
	port = ":50051"
)

var colaRetail=[]paquete{}
var colaPrioritario=[]paquete{}
var colaNormal=[]paquete{}
var registro map[int]orden

type server struct {
	pb.UnimplementedOrdenServiceServer
}

type orden struct {
	timestamp time.Time
	idPaquete string
	tipo string
	nombre string
	valor int
	origen string
	destino string
	seguimiento int
}
func newOrden(tipo string,nombre string,valor int,origen string, destino string,id string) *orden{
	ordenNueva := orden{nombre: nombre,tipo: tipo,valor: valor ,origen: origen,idPaquete: id, destino: destino}
	random := rand.NewSource(time.Now().UnixNano())
	ordenNueva.seguimiento=(rand.New(random)).Intn(492829)
	ordenNueva.timestamp=time.Now()
	return &ordenNueva
}
type paquete struct {
	idPaquete string
	tipo string
	valor int
	seguimiento int
	intentos int
	estado string
}

func newPaquete(idPaquete string, tipo string, valor int) *paquete{
	paqueteNuevo := paquete{idPaquete: idPaquete, tipo: tipo,valor: valor}
	random := rand.NewSource(time.Now().UnixNano())
	paqueteNuevo.seguimiento=(rand.New(random)).Intn(492829)
	paqueteNuevo.intentos= 0;
	paqueteNuevo.estado="en bodega"
	return &paqueteNuevo
}
func buscarPaquete(seguimiento int) orden{
	return registro[seguimiento]
}

func recibir(mensaje orden) orden{
	nuevaOrden :=newOrden(mensaje.tipo,mensaje.nombre,mensaje.valor,mensaje.origen,mensaje.destino,mensaje.idPaquete)
	//Aqui enviar nuevaOrden.seguimiento a cliente()
	nuevoPaquete := newPaquete(mensaje.idPaquete,mensaje.tipo,mensaje.valor)
	if nuevoPaquete.tipo=="retail"{
		colaRetail=append(colaRetail,*nuevoPaquete)
	} else if nuevoPaquete.tipo=="normal"{
		colaNormal=append(colaNormal,*nuevoPaquete)
	} else{
		colaPrioritario=append(colaPrioritario, *nuevoPaquete)
	}
	registro[nuevaOrden.seguimiento]= *nuevaOrden
	return *nuevaOrden
}

func enviarColas(){
	
}

func (s* server)replyToOrder(ctx context.Context,pedido *pb.SendToOrden) (*pb.ReplyFromOrden,error){
	orden := *newOrden(pedido.Tipo,pedido.Nombre,int(pedido.Valor),pedido.Origen,pedido.Destino,pedido.IdPaquete)
	orden=recibir(orden)
	seguimiento := pb.ReplyFromOrden{Seguimiento:int64(orden.seguimiento)}
	return &seguimiento,nil
}


func main() { 
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrdenServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

