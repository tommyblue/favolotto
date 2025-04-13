<template>
  <div
    class="bg-gray-800 min-h-screen flex flex-col items-center justify-center px-6 sm:px-8 lg:px-12 py-10 bg-gray-100">
    <img src="./assets/favolotto.png" alt="Favolotto" class="w-24 h-24 mb-6" />
    <h1 class="text-4xl font-sans font-bold text-indigo-400 mb-10 text-center">Favolotto</h1>

    <div class="w-full max-w-lg md:max-w-2xl lg:max-w-4xl bg-white p-8 rounded-lg shadow-lg">
      <h2 class="text-2xl font-semibold mb-6 text-gray-700">Add a new song</h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label for="nfcTag" class="block text-sm font-medium text-gray-700 mb-1">NFC Tag</label>
          <input v-model="nfcTag" type="text" id="nfcTag" placeholder="Insert NFC Tag"
            class="w-full px-4 py-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500" />
        </div>

        <div>
          <label for="songFile" class="block text-sm font-medium text-gray-700 mb-1">Select an MP3 file</label>
          <input type="file" id="songFile" @change="handleFileUpload" accept="audio/mp3"
            class="w-full px-4 py-2 border border-gray-300 rounded-md bg-white cursor-pointer" />
        </div>
      </div>

      <button @click="uploadSong"
        class="w-full bg-indigo-600 text-white font-semibold py-3 rounded-md mt-6 hover:bg-indigo-700 transition">
        Upload Song
      </button>

      <h4 class="text-xl mt-8 font-semibold text-indigo-300">Last read tag: {{ currentTag }}</h4>
    </div>

    <div class="mt-12 w-full max-w-lg md:max-w-2xl lg:max-w-4xl">
      <h2 class="text-2xl font-semibold text-indigo-300 mb-6 text-center">Songs list</h2>

      <table
        class="w-full border-collapse border border-gray-400 bg-white text-sm dark:border-gray-500 dark:bg-gray-800">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th
              class="w-1/3 border border-gray-300 p-4 text-left font-semibold text-gray-900 dark:border-gray-600 dark:text-gray-200">
              NFC tag</th>
            <th
              class="w-1/3 border border-gray-300 p-4 text-left font-semibold text-gray-900 dark:border-gray-600 dark:text-gray-200">
              Name</th>
            <th
              class="w-1/3 border border-gray-300 p-4 text-left font-semibold text-gray-900 dark:border-gray-600 dark:text-gray-200">
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="song in songs" :key="song.nfc_tag">
            <td class="border border-gray-300 p-4 text-gray-500 dark:border-gray-700 dark:text-gray-400">{{ song.nfc_tag
            }}</td>
            <td class="border border-gray-300 p-4 text-gray-500 dark:border-gray-700 dark:text-gray-400">{{ song.name }}
            </td>
            <td class="border border-gray-300 p-4 text-gray-500 dark:border-gray-700 dark:text-gray-400">
              <button @click="openModal(song.nfc_tag, song.name)"
                class="cursor-pointer text-red-400 font-semibold px-3 py-2 rounded-md hover:text-red-800 transition">
                Delete
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <TransitionRoot appear :show="alertOpen" as="template">
      <Dialog as="div" class="relative z-10" @close="closeAlert">
        <div class="fixed inset-0 bg-black bg-opacity-50" aria-hidden="true"></div>

        <div class="fixed inset-0 flex items-center justify-center px-6">
          <DialogPanel class="w-full max-w-sm md:max-w-md lg:max-w-lg bg-white rounded-lg p-8 shadow-xl">
            <DialogDescription class="mt-4 text-gray-600">
              Please add a file to upload and an NFC tag to associate with the song.
            </DialogDescription>
          </DialogPanel>
        </div>
      </Dialog>
    </TransitionRoot>

    <TransitionRoot appear :show="modalOpen" as="template">
      <Dialog as="div" class="relative z-10" @close="closeModal">
        <div class="fixed inset-0 bg-black bg-opacity-50" aria-hidden="true"></div>

        <div class="fixed inset-0 flex items-center justify-center px-6">
          <DialogPanel class="w-full max-w-sm md:max-w-md lg:max-w-lg bg-white rounded-lg p-8 shadow-xl">
            <DialogTitle class="text-xl font-bold text-gray-700">Delete the song?</DialogTitle>
            <DialogDescription class="mt-4 text-gray-600">
              Do you really want to delete the song "{{ currentSongToDelete }}" (tag "{{ currentTagToDelete }}")?
            </DialogDescription>

            <div class="mt-6 flex justify-end gap-4">
              <button @click="deleteSong(currentTagToDelete)"
                class="bg-red-600 text-white px-5 py-2.5 rounded-md hover:bg-red-700 transition">
                Yes, delete
              </button>
              <button @click="closeModal"
                class="bg-gray-300 text-black px-5 py-2.5 rounded-md hover:bg-gray-400 transition">
                Cancel
              </button>
            </div>
          </DialogPanel>
        </div>
      </Dialog>
    </TransitionRoot>
  </div>
</template>

<script>
import { usePolling } from "@/composables/usePolling";
import { Dialog, DialogDescription, DialogPanel, DialogTitle, TransitionRoot } from '@headlessui/vue';
import axios from 'axios';
import { onMounted, ref } from 'vue';

export default {
  components: {
    Dialog,
    DialogPanel,
    DialogTitle,
    DialogDescription,
    TransitionRoot,
  },
  setup() {
    const songs = ref([]);
    const nfcTag = ref('');
    const songFile = ref(null);
    const modalOpen = ref(false);
    const alertOpen = ref(false);
    const currentSongToDelete = ref(null);
    const currentTagToDelete = ref(null);
    const currentTag = ref('');

    const fetchData = async () => {
      const response = await fetch("/api/v1/tags/current");
      const data = await response.json();
      currentTag.value = data.nfc_tag;
    };

    usePolling(fetchData, 5000);

    console.log("currentTag", currentTag);
    const fetchSongs = async () => {
      try {
        const response = await axios.get('/api/v1/songs');
        songs.value = response.data;
      } catch (error) {
        console.error('Errore nel recupero delle canzoni:', error);
      }
    };

    const handleFileUpload = (event) => {
      songFile.value = event.target.files[0];
    };

    const uploadSong = async () => {
      if (!nfcTag.value || !songFile.value) {
        alertOpen.value = true;
        return;
      }
      const formData = new FormData();
      formData.append('nfc_tag', nfcTag.value);
      formData.append('song', songFile.value);
      try {
        await axios.put('/api/v1/song', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
        fetchSongs();
        nfcTag.value = '';
        songFile.value = null;
      } catch (error) {
        console.error('Errore nel caricamento della canzone:', error);
      }
    };

    const deleteSong = async (tag) => {
      try {
        await axios.delete('/api/v1/song', {
          data: { nfc_tag: tag },
        });
        fetchSongs();
        closeModal();
      } catch (error) {
        console.error("Errore nell'eliminazione della canzone:", error);
      }
    };

    const openModal = (tag, name) => {
      currentTagToDelete.value = tag;
      currentSongToDelete.value = name;
      modalOpen.value = true;
    };

    const closeModal = () => {
      modalOpen.value = false;
      currentTagToDelete.value = null;
      currentSongToDelete.value = null;
    };

    const closeAlert = () => {
      alertOpen.value = false;
    };

    onMounted(fetchSongs);

    return {
      currentTag,
      songs,
      nfcTag,
      songFile,
      modalOpen,
      alertOpen,
      closeAlert,
      currentSongToDelete,
      currentTagToDelete,
      fetchSongs,
      uploadSong,
      deleteSong,
      openModal,
      closeModal,
      handleFileUpload,
    };
  },
};
</script>

<style computed>
@import "tailwindcss";
</style>
