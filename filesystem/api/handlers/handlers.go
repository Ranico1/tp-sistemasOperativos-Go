package handlers

import (
	"fmt"
	"net/http"

	global "github.com/sisoputnfrba/tp-golang/filesystem/global"
	internal "github.com/sisoputnfrba/tp-golang/filesystem/internal"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	metodosHttp "github.com/sisoputnfrba/tp-golang/utils/metodosHttp"
)

type PeticionDump struct {
	NombreArchivo string `json:"nombre_archivo"`
	Tamanio       int    `json:"tamanio"`
	Contenido     []byte `json:"contenido"`
}

// 	TODO: Logear cuando accedo al bloque de indice, y cuando escribo el contenido en los otros bloques

func CrearArchivoDump(w http.ResponseWriter, r *http.Request) {

	// 1) Verificar si se cuenta con espacio disponible

	var peticion PeticionDump

	
	err := metodosHttp.DecodeHTTPBody(r, &peticion)
	if err != nil {
		global.Logger.Log("Error al decodificar el body de la peticion: "+err.Error(), log.ERROR)
		http.Error(w, "Error al decodificar el body de la peticion", http.StatusNoContent)
		return
	}

	global.Logger.Log(fmt.Sprintf("NombreArchivo: %s, Tamanio: %d, Contenido: %v", peticion.NombreArchivo, peticion.Tamanio, peticion.Contenido), log.DEBUG)
	
	//fmt.Println("CONTENIDO :", peticion.Contenido)

	if !internal.EspacioDisponible(peticion.Tamanio) {
		global.Logger.Log("No hay espacio disponible para el archivo", log.ERROR)
		http.Error(w, "No hay espacio disponible para el archivo", http.StatusNoContent)
		return
	}

	//2) Reservar espacio en el bitmap

	bloqueIndice, punterosBloques, err := internal.ReservarBloques(peticion.Tamanio, peticion.NombreArchivo)
	if err != nil {
		global.Logger.Log("Error al reservar bloques", log.ERROR)
		http.Error(w, "Error al escribir contenido", http.StatusNoContent)
		return
	}

	//3) Crear archivo de metadata que sea del tipo Archivo

	err = internal.CrearArchivoMetadata(peticion.NombreArchivo, bloqueIndice, peticion.Tamanio)
	if err != nil {
		global.Logger.Log("Error al crear archivo metadata", log.ERROR)
		http.Error(w, "Error en la creacion del archivo metadata", http.StatusNoContent)
	}

	// 4) Abrir archivo de bloques.dat para luego escribirlo con los punteros
	archivoBloques, err := internal.AbrirArchivoBloquesDat()
	if err != nil {
		global.Logger.Log("Error al abrir bloques.dat", log.ERROR)
		http.Error(w, "Error al abrir el archivo de bloques.dat", http.StatusNoContent)
	}

	// 5) Acceder al bloque de indice y grabar los punteros a cada bloque de datos

	err = internal.GrabarPunterosEnBloqueIndice(bloqueIndice, peticion.Tamanio, punterosBloques, peticion.NombreArchivo, archivoBloques)
	if err != nil {
		global.Logger.Log("Error al cargar punteros en el bloque de indice", log.ERROR)
		http.Error(w, "Error en la carga de punteros en el bloque de indice", http.StatusNoContent)
	}

	// 6) Con los punteros escribir el contenido en los bloques de datos

	err = internal.GrabarContenidoEnBloques(punterosBloques, peticion.Contenido, peticion.NombreArchivo, archivoBloques)
	if err != nil {
		global.Logger.Log("Error al grabar el contenido en los bloques de datos", log.ERROR)
		http.Error(w, "Error al escribir contenido", http.StatusNoContent)
		return
	}

	// 7) Cargar slice Bitmap

	internal.CargarSliceABitmap()

	// 8) Cerrar archivo bloques.dat

	internal.CerrarArchivoBloquesDat(archivoBloques)


	// 9) Logear fin de peticion

	global.Logger.Log(fmt.Sprintf("## Fin solicitud - Archivo: %s", peticion.NombreArchivo), log.INFO)

	w.WriteHeader(http.StatusOK)

}
