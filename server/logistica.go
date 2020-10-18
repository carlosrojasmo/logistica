package main
import (
	"time"
	"math/rand"
	"log"
	"net"
	"context"
	"google.golang.org/grpc"
	pb "../proto"
	"fmt"
)
const (
	port = ":50051"
	port2 = ":50054"
)
var paqueteAux=newPaquete("400","null",1) 
var colaRetail=[]paquete{}
var colaPrioritario=[]paquete{}
var colaNormal=[]paquete{}
var registro= make(map[int]orden)
var registroPaquete= make (map[int]paquete)

type server struct {
	pb.UnimplementedOrdenServiceServer
}
type serverDos struct {
	pb.UnimplementedCamionDeliveryServer
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
func newOrden(tipo string,nombre string,valor int,origen string, destino string,id string) orden{
	ordenNueva := orden{nombre: nombre,tipo: tipo,valor: valor ,origen: origen,idPaquete: id, destino: destino}
	random := rand.NewSource(time.Now().UnixNano())
	ordenNueva.seguimiento=(rand.New(random)).Intn(492829)
	ordenNueva.timestamp=time.Now()
	return ordenNueva
}
type paquete struct {
	idPaquete string
	tipo string
	valor int
	seguimiento int
	intentos int
	estado string
}

func newPaquete(idPaquete string, tipo string, valor int) paquete{
	paqueteNuevo := paquete{idPaquete: idPaquete, tipo: tipo,valor: valor}
	random := rand.NewSource(time.Now().UnixNano())
	paqueteNuevo.seguimiento=(rand.New(random)).Intn(492829)
	paqueteNuevo.intentos= 0;
	paqueteNuevo.estado="Recibido"
	return paqueteNuevo
}
func buscarPaquete(seguimiento int) paquete{
	return registroPaquete[seguimiento]
}

func recibir(mensaje orden) orden{
	nuevaOrden :=newOrden(mensaje.tipo,mensaje.nombre,mensaje.valor,mensaje.origen,mensaje.destino,mensaje.idPaquete)
	fmt.Println(nuevaOrden)
	nuevoPaquete := newPaquete(mensaje.idPaquete,mensaje.tipo,mensaje.valor)
	registroPaquete[nuevoPaquete.seguimiento]=nuevoPaquete
	if nuevoPaquete.tipo=="retail"{
		colaRetail=append(colaRetail,nuevoPaquete)
	} else if nuevoPaquete.tipo=="normal"{
		colaNormal=append(colaNormal,nuevoPaquete)
	} else{
		colaPrioritario=append(colaPrioritario, nuevoPaquete)
	}
	registro[nuevaOrden.seguimiento]= nuevaOrden
	return nuevaOrden
}

func enviarColas(tipo string) paquete{
	var paquetePedido paquete
	if tipo=="retail"{
		if len(colaRetail)==0{
			if len (colaPrioritario)==0{
				paquetePedido=paqueteAux
			}else{
				 paquetePedido = colaPrioritario[0]
				colaPrioritario=colaPrioritario[1:]
			}
		}else{
			 paquetePedido =colaRetail[0]
			colaRetail=colaRetail[1:]
		}
	}else{
		if len(colaPrioritario)==0{
			if len(colaNormal)==0{
				paquetePedido= paqueteAux
			}else{
				 paquetePedido =colaNormal[0]
				colaNormal=colaNormal[1:]
			}
		}else{
			paquetePedido =colaPrioritario[0]
			colaPrioritario=colaPrioritario[1:]
		}
	}
	
	return paquetePedido
}

func recibirReporte(idPaquete string,entregado bool,intentos int64) string{
	for seguimiento,paquete:=range registroPaquete{
		if paquete.idPaquete==idPaquete{
			paquete.estado="No entregado"
			paquete.intentos=int(intentos)
			registroPaquete[seguimiento]=paquete
		}
	}
	return "ok"
}

func (s* server)ReplyToOrder(ctx context.Context,pedido *pb.SendToOrden) (*pb.ReplyFromOrden,error){
	orden := newOrden(pedido.Tipo,pedido.Nombre,int(pedido.Valor),pedido.Origen,pedido.Destino,pedido.IdPaquete)
	orden=recibir(orden)
	seguimiento := pb.ReplyFromOrden{Seguimiento:int64(orden.seguimiento)}
	return &seguimiento,nil
}
func (s* server)GetState(ctx context.Context, seguimiento *pb.ReplyFromOrden) (*pb.InfoSeguimiento, error){
	paq:=buscarPaquete(int(seguimiento.Seguimiento))
	fmt.Println(seguimiento.Seguimiento)
	fmt.Println(paq)
	estado:=pb.InfoSeguimiento{Estado:paq.estado}

	return &estado,nil
}

func (s* server)GetPack(ctx context.Context, pedido *pb.AskForPack) (*pb.SendPack, error){
	tipo := pedido.Tipo
	paqueteEncontrado := enviarColas(tipo)
	ordenPaquete:= registro[paqueteEncontrado.seguimiento]
	paqueteEnviado := pb.SendPack{IdPaquete: ordenPaquete.idPaquete,Tipo:ordenPaquete.tipo,Nombre:ordenPaquete.nombre,Valor:int64(ordenPaquete.valor),Origen:ordenPaquete.origen,Destino:ordenPaquete.destino}
	return  &paqueteEnviado,nil
}

func (s* server)Report(ctx context.Context, reporte *pb.ReportDelivery,) (*pb.ReportOk, error){
	idPaquete:= reporte.IdPaquete
	entregado:= reporte.Entregado
	intentos:= reporte.Intentos
	resultado := recibirReporte(idPaquete,entregado,intentos)
	reporteok:= pb.ReportOk{Ok: resultado}
	return &reporteok,nil
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

	
	lis2, err2 := net.Listen("tcp", port2)
	if err2 != nil {
		log.Fatalf("failed to listen: %v", err2)
	}
	s2 := grpc.NewServer()
	pb.RegisterCamionDeliveryServer(s2, &serverDos{})
	if err2 := s2.Serve(lis2); err2 != nil {
		log.Fatalf("failed to serve: %v", err2)
	}
}

