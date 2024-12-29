package internal

import (
	//"bytes"
	//"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"unsafe"

	global "github.com/sisoputnfrba/tp-golang/filesystem/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)



// Funcion que recorre el bitmap y cuando encuentra espacio para los bloques solicitados y el indice, devuelve true
func EspacioDisponible(tamanio int) bool {

	// Calculo los bloques necesarios para el archivo, redondeando para arriba por el bloque de indice
	bloquesNecesarios := tamanio / global.FSConfig.BlockSize

	resto := tamanio % global.FSConfig.BlockSize

	if resto != 0 {
		bloquesNecesarios++
	}

	//fmt.Println("TAMANIO", tamanio)
	//fmt.Println("BLOQUES NECESARIOS: ", bloquesNecesarios)

	tamanioMaximo := int(unsafe.Sizeof(uint32(0))) //Tamanio maximo que tiene un archivo, ya que tiene 1 solo bloque de indice

	//fmt.Println("BLOQUES MAXIMOS EN EL INDICE", global.FSConfig.BlockSize/tamanioMaximo)

	//Verifico que todos los bloques entren en el bloque de indice
	if bloquesNecesarios > global.FSConfig.BlockSize/tamanioMaximo {

		return false
	}

	bloquesDisponibles := 0
	for i := 0; i < len(global.Bitmap); i++ {
		for j := 0; j < 8; j++ {
			if (global.Bitmap[i] & (1 << j)) == 0 {
				bloquesDisponibles++
				if bloquesDisponibles >= bloquesNecesarios+1 { // +1 para el bloque de índice
					return true
				}
			}
		}
	}

	return false
}

// Funcion que reserva los bloques necesarios para el archivo y devuelve el bloque de indice y los punteros a los bloques de datos
func ReservarBloques(tamanio int, nombreArchivo string) (int, []int, error) {
	bloquesNecesarios := tamanio / global.FSConfig.BlockSize

	resto := tamanio % global.FSConfig.BlockSize

	if resto != 0 {
		bloquesNecesarios++
	}

	punterosBloqueIndice := []int{}
	

	bloqueIndice, err := AsignarBloque(nombreArchivo)
	if err != nil {
		return -1, punterosBloqueIndice, err
	}

	//Reservo los bloques de datos necesarios

	for i := 0; i < bloquesNecesarios; i++ {
		bloque, err := AsignarBloque(nombreArchivo)
		if err != nil {
			return -1, punterosBloqueIndice, err //devuelvo basura, igualmente en el handlers se maneja el error
		}
		punterosBloqueIndice = append(punterosBloqueIndice, bloque)
	}
	
	global.Logger.Log(fmt.Sprintf("PunterosBloques: %v", punterosBloqueIndice), log.DEBUG)

	return bloqueIndice, punterosBloqueIndice, nil

}

// Esta funcion busca el primer bloque libre para asignar
func AsignarBloque(nombreArchivo string) (int, error) {
	var err error
	var numeroBloque int
	for i := 0; i < len(global.Bitmap); i++ {
		for j := 0; j < 8; j++ {
			if (global.Bitmap[i] & (1 << j)) == 0 { //Si el bit no esta ocupado:

				global.Bitmap[i] |= (1 << j) //Marcar el BIT como ocupado

				numeroBloque = i*8 + j //Cada byte son 8 bits (i*8+j) es el numero de bloque

				global.Logger.Log(fmt.Sprintf("## Bloque asignado: %d - Archivo: %s - Bloques Libres: %d", numeroBloque, nombreArchivo, CantidadBloquesLibres()), log.INFO)

				return numeroBloque, nil
			}
		}
	}
	return -1, err
}

func CantidadBloquesLibres() int {
	bloquesLibres := 0
	for i := 0; i < len(global.Bitmap); i++ {
		for j := 0; j < 8; j++ {
			if (global.Bitmap[i] & (1 << j)) == 0 {
				bloquesLibres++
			}
		}
	}
	return bloquesLibres
}

func GrabarPunterosEnBloqueIndice(bloqueIndice int, tamanio int, punterosBloques []int, nombreArchivo string, archivoBloques *os.File) error {

	// Calculo la posicion del bloque indice
	posicionIndice := int64(bloqueIndice * global.FSConfig.BlockSize)

	global.Logger.Log(fmt.Sprintf("## Acceso Bloque - Archivo: %s - Tipo Bloque: INDICE - Bloque File System: %d", nombreArchivo, bloqueIndice), log.INFO)
	time.Sleep(time.Duration(global.FSConfig.BlockAccessDelay) * time.Millisecond)
	
	// Escribir los punteros en el bloque de índice

	//Me posiciono en archivoBloques en la posicion del bloque de indice
	_, err := archivoBloques.Seek(posicionIndice, 0) //0 representa SEEK_SET (comienzo del archivo)

	if err != nil {
		global.Logger.Log("Error al posicionar el archivo: "+err.Error(), log.ERROR)
		return err
	}

	contenidoParaEscribir := punterosABytes(punterosBloques)

	// Escribo los punteros en el bloque de indice
	_, err = archivoBloques.Write(contenidoParaEscribir)
	if err != nil {
		global.Logger.Log("Error al escribir el archivo: "+err.Error(), log.ERROR)
		return err
	}


	return nil
}

func punterosABytes(bloque []int) []byte {
	byteSlice := make([]byte, len(bloque))
	for i, v := range bloque {
		byteSlice[i] = byte(v)
	}
	return byteSlice
}

func CrearArchivoMetadata(nombreArchivo string, bloqueIndice int, tamanio int) error {
	var archivoMetadata global.Archivo

	var err error
	ruta := global.FSConfig.MountDir + "/files/"

	archivo, err := os.Create(ruta + nombreArchivo)
	if err != nil {
		global.Logger.Log("Error al crear el archivo: "+err.Error(), log.ERROR)
		return err
	}


	// 3.1) Asignarle a archivoMetadata el bloque de indice y el tamanio, y luego codificarlo en el archivo

	archivoMetadata.IndexBlock = bloqueIndice
	archivoMetadata.Size = tamanio

	encoder := json.NewEncoder(archivo)
	err = encoder.Encode(archivoMetadata)
	if err != nil {
		global.Logger.Log("Error al codificar el archivo: "+err.Error(), log.ERROR)
		return err
	}

	global.Logger.Log(fmt.Sprintf("## Archivo Creado: %s  - Tamaño: %d ", nombreArchivo, tamanio), log.INFO)

	defer archivo.Close()

	return nil
}

func GrabarContenidoEnBloques(punterosBloques []int, contenido []byte, nombreArchivo string, archivoBloques *os.File) error {

	// En cada iteracion calculo la posicion segun el numero de bloque

	fmt.Println("punterosBloques: ", punterosBloques)
	for i, puntero := range punterosBloques {

		posicionBloque := int64(puntero * global.FSConfig.BlockSize)


		//Me posiciono en archivoBloques en la posicion del bloque de indice
		_, err := archivoBloques.Seek(posicionBloque, 0) //0 representa SEEK_SET (comienzo del archivo)

		if err != nil {
			global.Logger.Log("Error al posicionar el archivo: "+err.Error(), log.ERROR)
			return err
		}

		global.Logger.Log(fmt.Sprintf("## Acceso Bloque - Archivo: %s - Tipo Bloque: DATOS - Bloque File System: %d", nombreArchivo, puntero), log.INFO)
		time.Sleep(time.Duration(global.FSConfig.BlockAccessDelay) * time.Millisecond)

		// Calculo el fragmento del contenido a escribir en este bloque
		inicio := i * global.FSConfig.BlockSize
		fin := inicio + global.FSConfig.BlockSize
		if inicio >= len(contenido) {
			return nil
		}
		if fin > len(contenido) {
			fin = len(contenido)
		}
		fragmento := contenido[inicio:fin]
		

		// Escribo el fragmento del contenido que le corresponde a ese bloque
		_, err = archivoBloques.Write(fragmento)

		if err != nil {
			global.Logger.Log("Error al escribir el archivo: "+err.Error(), log.ERROR)
			return err
		}

	}

	return nil
}

func AbrirArchivoBloquesDat() (*os.File, error) {

	var err error

	ruta := global.FSConfig.MountDir + "/bloques.dat"

	tamanio := global.FSConfig.BlockCount * global.FSConfig.BlockSize

	archivoBloques, err := os.OpenFile(ruta, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		global.Logger.Log(fmt.Sprint("Error al crear/abrir archivo: ", err), log.ERROR)
		return archivoBloques, err
	}

	err = archivoBloques.Truncate(int64(tamanio))
	if err != nil {
		global.Logger.Log(fmt.Sprint("Error al ajustar el archivo: ", err), log.ERROR)
		return archivoBloques, err
	}

	return archivoBloques , nil
}



func CerrarArchivoBloquesDat(archivoBloques *os.File) {
	archivoBloques.Close()
}

func CargarSliceABitmap() {
	_, err := global.ArchivoBitmap.Seek(0, 0)
	if err != nil {
		global.Logger.Log("error al mover el puntero del archivo:"+err.Error(), log.ERROR)
	}

	_, err = global.ArchivoBitmap.Write(global.Bitmap)
	if err != nil {
		global.Logger.Log("Error al escribir en el archivo Bitmap: "+err.Error(), log.ERROR)
		return
	}

	err = global.ArchivoBitmap.Sync()
	if err != nil {
		global.Logger.Log("Error al sincronizar Bitmap: "+err.Error(), log.ERROR)
		return
	}

	global.Logger.Log("Contenido del Bitmap escrito en el archivo exitosamente.", log.INFO)

}
