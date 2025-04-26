import { Dialog, DialogDescription, DialogPanel, DialogTitle } from "@headlessui/react"
import React, { useEffect, useState } from "react"
import { toast } from "sonner"
import { Song } from "../types"

export default function () {
	const [songs, setSongs] = useState<Song[]>([])
	const [modalOpen, setModalOpen] = useState(false)
	const [songToDelete, setSongToDelete] = useState<Song | null>(null)
	const [isDeletingSong, setIsDeletingSong] = useState(false)

	function openModal(song: Song) {
		setSongToDelete(song)
		setModalOpen(true)
	}
	function closeModal() {
		setSongToDelete(null)
		setModalOpen(false)
	}

	function deleteSong() {
		if (!songToDelete) {
			return
		}

		setIsDeletingSong(true)

		fetch("/api/v1/song", {
			method: "DELETE",
			headers: {
				"content-type": "application/json",
			},
			body: JSON.stringify({ nfc_tag: songToDelete.nfc_tag }),
		})
			.then(fetchSongs)
			.catch(e => toast.error(`cannot delete song: ${e}`))
			.finally(() => {
				setIsDeletingSong(false)
				setModalOpen(false)
			})
	}

	useEffect(() => {
		const interval = setInterval(fetchSongs, 5000)
		fetchSongs()
		return () => clearInterval(interval)
	}, [])

	function fetchSongs() {
		fetch("/api/v1/songs")
			.then(response => response.json())
			.then(songs => setSongs(songs))
			.catch(e => {
				toast.error(`cannot fetch: ${e}`)
			})
	}

	return (
		<>
			<div className="mt-12 w-full max-w-lg md:max-w-2xl lg:max-w-4xl">
				<h2 className="text-2xl font-semibold text-indigo-300 mb-6 text-center">Songs list</h2>

				<table className="w-full border-collapse border border-gray-400 bg-white text-sm dark:border-gray-500 dark:bg-gray-800">
					<thead className="bg-gray-50 dark:bg-gray-700">
						<tr>
							<th className="w-1/3 border border-gray-300 p-4 text-left font-semibold text-gray-900 dark:border-gray-600 dark:text-gray-200">
								NFC tag
							</th>
							<th className="w-1/3 border border-gray-300 p-4 text-left font-semibold text-gray-900 dark:border-gray-600 dark:text-gray-200">
								Name
							</th>
							<th className="w-1/3 border border-gray-300 p-4 text-left font-semibold text-gray-900 dark:border-gray-600 dark:text-gray-200"></th>
						</tr>
					</thead>
					<tbody>
						{songs.map(song => (
							<tr key={song.nfc_tag}>
								<td className="border border-gray-300 p-4 text-gray-500 dark:border-gray-700 dark:text-gray-400">
									{song.nfc_tag}
								</td>
								<td className="border border-gray-300 p-4 text-gray-500 dark:border-gray-700 dark:text-gray-400">
									{song.name}
								</td>
								<td className="border border-gray-300 p-4 text-gray-500 dark:border-gray-700 dark:text-gray-400">
									<button
										disabled={isDeletingSong}
										onClick={_ => openModal(song)}
										className="cursor-pointer text-red-400 font-semibold px-3 py-2 rounded-md hover:text-red-800 transition"
									>
										Delete
									</button>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
			<Dialog as="div" className="relative z-10" open={modalOpen} onClose={() => closeModal()}>
				<div className="fixed inset-0 bg-black bg-opacity-50" aria-hidden="true"></div>
				<div className="fixed inset-0 flex items-center justify-center px-6">
					<DialogPanel className="w-full max-w-sm md:max-w-md lg:max-w-lg bg-white rounded-lg p-8 shadow-xl">
						<DialogTitle className="text-xl font-bold text-gray-700">Delete the song?</DialogTitle>
						<DialogDescription className="mt-4 text-gray-600">
							Do you really want to delete the song "{songToDelete?.name}" (tag "{songToDelete?.nfc_tag}")?
						</DialogDescription>

						<div className="mt-6 flex justify-end gap-4">
							<button
								onClick={() => deleteSong()}
								className="bg-red-600 text-white px-5 py-2.5 rounded-md hover:bg-red-700 transition"
							>
								Yes, delete
							</button>
							<button
								onClick={() => closeModal()}
								className="bg-gray-300 text-black px-5 py-2.5 rounded-md hover:bg-gray-400 transition"
							>
								Cancel
							</button>
						</div>
					</DialogPanel>
				</div>
			</Dialog>
		</>
	)
}
