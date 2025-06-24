package main

import (
	"context"
	"fmt"
	"html/template" // Digunakan untuk merender HTML
	"log"
	"net/http"

	"google.golang.org/api/option"    // Untuk konfigurasi API Key
	"google.golang.org/api/sheets/v4" // Untuk berinteraksi dengan Google Sheets API
)

// Peserta merepresentasikan struktur satu baris data dari spreadsheet SCC MLBB.
// PENTING: Urutan field di struct ini HARUS sesuai dengan urutan kolom di spreadsheet Anda.
// Kolom: Timestamp (A), Nomor Pendaftaran (B), Token (C), Email Address (D), Nama Tim (E),
// Nama Lengkap Leader (F), Nomor WhatsApp Aktif Leader (G), Email Aktif Leader (H),
// Asal Daerah (I), Anggota Tim (J), Unggah Logo Tim (K), Pernyataan Persetujuan (L)
type Peserta struct {
	Timestamp                string
	NomorPendaftaran         string
	Token                    string
	EmailAddress             string
	NamaTim                  string
	NamaLengkapLeader        string
	NomorWhatsAppAktifLeader string
	EmailAktifLeader         string
	AsalDaerah               string
	AnggotaTim               string
	UnggahLogoTim            string // URL logo jika diunggah via G-Form ke G-Drive
	PernyataanPersetujuan    string
}

// PageData adalah struktur yang akan kita kirimkan ke template HTML.
type PageData struct {
	PesertaList []Peserta // Mengganti Usulans menjadi PesertaList
}

func main() {
	// --- PENTING: GANTI NILAI INI DENGAN DATA ANDA YANG SEBENARNYA ---
	const API_KEY = "AIzaSyC7bTd3vDlXJ2Xv5_HbSmU2Y6BZXOwBIWI"             // API Key Anda dari Google Cloud Console
	const SPREADSHEET_ID = "117LG8mUa9ZM5_YGpisp2-2Pjiz2PAsvyOn-4amz1nU0" // Spreadsheet ID baru Anda
	const SHEET_NAME = "DataMasterPendaftaranSCCMLBB-2025"                // Nama sheet/tab di spreadsheet Anda
	// --- AKHIR PENGGANTIAN ---

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		srv, err := sheets.NewService(ctx, option.WithAPIKey(API_KEY))
		if err != nil {
			log.Printf("Kesalahan saat membuat client Sheets API: %v", err)
			http.Error(w, "Maaf, ada masalah teknis saat menghubungkan ke Google Sheets.", http.StatusInternalServerError)
			return
		}

		// Range data yang ingin diambil: dari kolom A (Timestamp) hingga L (Pernyataan Persetujuan).
		// A2:L berarti mulai dari baris 2 (melewati header) hingga kolom L.
		readRange := fmt.Sprintf("%s!A2:L", SHEET_NAME) // Mengubah range dari J menjadi L

		resp, err := srv.Spreadsheets.Values.Get(SPREADSHEET_ID, readRange).Do()
		if err != nil {
			log.Printf("Kesalahan saat mengambil data dari spreadsheet: %v", err)
			http.Error(w, "Maaf, tidak bisa memuat data peserta dari spreadsheet.", http.StatusInternalServerError)
			return
		}

		var pesertaList []Peserta // Menggunakan slice Peserta

		if len(resp.Values) == 0 {
			fmt.Println("Tidak ada data peserta ditemukan di spreadsheet.")
		} else {
			// Kolom: A, B, C, D, E, F, G, H, I, J, K, L (total 12 kolom)
			// Indeks: 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11
			for _, row := range resp.Values {
				// Pastikan baris memiliki minimal 12 kolom
				if len(row) >= 12 {
					peserta := Peserta{
						Timestamp:                fmt.Sprintf("%v", row[0]),
						NomorPendaftaran:         fmt.Sprintf("%v", row[1]),
						Token:                    fmt.Sprintf("%v", row[2]),
						EmailAddress:             fmt.Sprintf("%v", row[3]),
						NamaTim:                  fmt.Sprintf("%v", row[4]),
						NamaLengkapLeader:        fmt.Sprintf("%v", row[5]),
						NomorWhatsAppAktifLeader: fmt.Sprintf("%v", row[6]),
						EmailAktifLeader:         fmt.Sprintf("%v", row[7]),
						AsalDaerah:               fmt.Sprintf("%v", row[8]),
						AnggotaTim:               fmt.Sprintf("%v", row[9]),
						UnggahLogoTim:            fmt.Sprintf("%v", row[10]),
						PernyataanPersetujuan:    fmt.Sprintf("%v", row[11]),
					}
					pesertaList = append(pesertaList, peserta)
				} else {
					log.Printf("Baris dilewati karena kolom tidak lengkap (kurang dari 12): %v", row)
				}
			}
		}

		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			log.Printf("Kesalahan saat memuat template HTML: %v", err)
			http.Error(w, "Maaf, halaman tidak dapat dimuat.", http.StatusInternalServerError)
			return
		}

		// Mengirim data ke template HTML
		dataForTemplate := PageData{
			PesertaList: pesertaList, // Mengganti Usulans menjadi PesertaList
		}

		err = tmpl.Execute(w, dataForTemplate)
		if err != nil {
			log.Printf("Kesalahan saat mengeksekusi template: %v", err)
			http.Error(w, "Maaf, ada masalah saat menampilkan data.", http.StatusInternalServerError)
		}
	})

	fmt.Println("Server aplikasi pendaftaran peserta SCC MLBB 2025 berjalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
