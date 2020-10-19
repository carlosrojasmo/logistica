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
	"github.com/streadway/amqp"
	"encoding/json"
)
const (
	port = ":50051"
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
/*type serverDos struct {
	pb.UnimplementedCamionDeliveryServer
}*/

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
	IDPaquete string
	Tipo string
	Valor int
	Seguimiento int
	Intentos int
	Estado string
}

func newPaquete(idPaquete string, tipo string, valor int) paquete{
	paqueteNuevo := paquete{IDPaquete: idPaquete, Tipo: tipo,Valor: valor}
	random := rand.NewSource(time.Now().UnixNano())
	paqueteNuevo.Seguimiento=(rand.New(random)).Intn(492829)
	paqueteNuevo.Intentos= 0;
	paqueteNuevo.Estado="En bodega"
	return paqueteNuevo
}
func buscarPaquete(seguimiento int) paquete{
	return registroPaquete[seguimiento]
}

func recibir(mensaje orden) orden{
	nuevaOrden :=newOrden(mensaje.tipo,mensaje.nombre,mensaje.valor,mensaje.origen,mensaje.destino,mensaje.idPaquete)
	nuevoPaquete := newPaquete(mensaje.idPaquete,mensaje.tipo,mensaje.valor)
	fmt.Println("Nueva Orden: ",nuevaOrden)
	fmt.Println("Nuevo Paquete: ",nuevoPaquete)
	registroPaquete[nuevoPaquete.Seguimiento]=nuevoPaquete
	if nuevoPaquete.Tipo=="retail"{
		colaRetail=append(colaRetail,nuevoPaquete)
	} else if nuevoPaquete.Tipo=="normal"{
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
	paqueteEdit:=registroPaquete[paquetePedido.Seguimiento]
	paqueteEdit.Estado="En camino"
	registroPaquete[paquetePedido.Seguimiento]=paqueteEdit
	return paquetePedido
}

func finanza(paquete paquete){
	conn, err := amqp.Dial("amqp://logistica:logistica@10.10.28.102:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	  )
	  failOnError(err, "Failed to declare a queue")
	  pedido,err:=json.Marshal(paquete)
	  body := string(pedido)
	  err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing {
		  ContentType: "text/plain",
		  Body:        []byte(body),
		})
	  failOnError(err, "Failed to publish a message")
}

func recibirReporte(idPaquete string,entregado bool,intentos int64) string{
	var paqueteReal paquete
	for seguimiento,paquete:=range registroPaquete{
		if paquete.IDPaquete==idPaquete{
			if entregado==true{
				paquete.Estado="Recibido"
			}else{
				paquete.Estado="No Recibido"
			}
			paquete.Intentos=int(intentos)
			registroPaquete[seguimiento]=paquete
			paqueteReal=paquete
		}
		
	}
	finanza(paqueteReal)
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
	estado:=pb.InfoSeguimiento{Estado:paq.Estado}

	return &estado,nil
}

func (s* server)GetPack(ctx context.Context, pedido *pb.AskForPack) (*pb.SendPack, error){
	tipo := pedido.Tipo
	paqueteEncontrado := enviarColas(tipo)
	ordenPaquete:= registro[paqueteEncontrado.Seguimiento]
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

func failOnError(err error, msg string) {
	if err != nil {
	  log.Fatalf("%s: %s", msg, err)
	}
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


