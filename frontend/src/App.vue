<script setup>
import axios from "axios";
import { onMounted, ref } from "vue";

const songs = ref([]);
    const nfcTag = ref("");
    const songFile = ref(null);

    const fetchSongs = async () => {
      try {
        const response = await axios.get("http://localhost:5173/api/v1/songs");
        songs.value = response.data;
      } catch (error) {
        console.error("Errore nel recupero delle canzoni:", error);
      }
    };

    const uploadSong = async () => {
      if (!nfcTag.value || !songFile.value) {
        alert("Inserisci un tag NFC e seleziona un file MP3");
        return;
      }
      const formData = new FormData();
      formData.append("nfc_tag", nfcTag.value);
      formData.append("file", songFile.value);
      try {
        await axios.put("/api/v1/song", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        fetchSongs();
        nfcTag.value = "";
        songFile.value = null;
      } catch (error) {
        console.error("Errore nel caricamento della canzone:", error);
      }
    };

    const deleteSong = async (tag) => {
      try {
        await axios.delete("/api/v1/song", {
          data: { nfc_tag: tag },
        });
        fetchSongs();
      } catch (error) {
        console.error("Errore nell'eliminazione della canzone:", error);
      }
    };

    onMounted(fetchSongs);
</script>

<template>
  <div>
      <h1>Gestione Canzoni NFC</h1>
      <div>
        <input v-model="nfcTag" placeholder="Inserisci NFC Tag" />
        <input type="file" @change="event => songFile = event.target.files[0]" accept="audio/mp3" />
        <button @click="uploadSong">Carica</button>
      </div>
      <ul>
        <li v-for="song in songs" :key="song.nfc_tag">
          {{ song.nfc_tag }}
          <button @click="deleteSong(song.nfc_tag)">Elimina</button>
        </li>
      </ul>
    </div>
</template>

<style scoped>
</style>
