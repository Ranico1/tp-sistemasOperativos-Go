package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    if len(os.Args) < 4 {
        fmt.Println("Uso: modify_json <MODULO> <CLAVE> <VALOR>")
        return
    }

    modulo := os.Args[1]
    clave := os.Args[2]
    valor := os.Args[3]
    carpeta := filepath.Join(modulo, "config")

    var newValue interface{}
    err := json.Unmarshal([]byte(valor), &newValue)
    if err != nil {
        newValue = valor
    }

    err = filepath.Walk(carpeta, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && filepath.Ext(path) == ".json" {
            file, err := os.ReadFile(path)
            if err != nil {
                fmt.Printf("Error al leer el archivo %s: %v\n", path, err)
                return nil
            }

            var data map[string]interface{}
            err = json.Unmarshal(file, &data)
            if err != nil {
                fmt.Printf("Error al parsear JSON en el archivo %s: %v\n", path, err)
                return nil
            }

            data[clave] = newValue

            newJSON, err := json.MarshalIndent(data, "", "  ")
            if err != nil {
                fmt.Printf("Error al serializar JSON en el archivo %s: %v\n", path, err)
                return nil
            }

            err = os.WriteFile(path, newJSON, 0644)
            if err != nil {
                fmt.Printf("Error al escribir el archivo %s: %v\n", path, err)
                return nil
            }

            fmt.Printf("El archivo %s ha sido modificado correctamente\n", path)
        }
        return nil
    })

    if err != nil {
        fmt.Printf("Error al buscar archivos en la carpeta %s: %v\n", carpeta, err)
    }
}