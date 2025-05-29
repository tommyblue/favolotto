import React, { useEffect, useState } from "react"
import { toast } from "sonner"

export default function() {
    const [currentTag, setCurrentTag] = useState("")
    const [nfcTagToUpload, setNfcTagToUpload] = useState<string>("")
    const [fileToUpload, setFileToUpload] = useState<File | null>(null)
    const [isUploading, setIsUploading] = useState(false)

    useEffect(() => {
        const interval = setInterval(fetchCurrentTag, 5000)
        fetchCurrentTag()
        return () => clearInterval(interval)
    }, [])

    type CurrentTagResponse = {
        nfc_tag: string
    }

    function fetchCurrentTag() {
        fetch("/api/v1/tags/current")
            .then<CurrentTagResponse>(response => response.json())
            .then(data => setCurrentTag(data.nfc_tag))
            .catch(e => toast.error(`cannot get current tag: ${e}`))
    }

    function uploadSong() {
        if (fileToUpload === null || nfcTagToUpload === "") {
            toast.warning("Please add a file to upload and an NFC tag to associate with the song.")

            return
        }

        setIsUploading(true)

        const formData = new FormData()
        formData.append("nfc_tag", nfcTagToUpload)
        formData.append("song", fileToUpload)

        fetch("/api/v1/song", {
            method: "PUT",
            body: formData,
        })
            .catch(e => toast.error(`cannot upload: ${e}`))
            .finally(() => setIsUploading(false))
    }

    return (
        <>
            <div className="w-full max-w-lg md:max-w-2xl lg:max-w-4xl bg-white p-8 rounded-lg shadow-lg">
                <h2 className="text-2xl font-semibold mb-6 text-gray-700">Add a new song</h2>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">NFC Tag</label>
                        <input
                            type="text"
                            id="nfcTag"
                            placeholder="Insert NFC Tag"
                            className="w-full px-4 py-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
                            onChange={event => {
                                setNfcTagToUpload(event.target.value)
                            }}
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Select an MP3 file</label>
                        <input
                            type="file"
                            id="songFile"
                            onChange={event => {
                                const target = event.target as HTMLInputElement
                                target.files && setFileToUpload(target.files[0])
                            }}
                            accept="audio/mp3"
                            className="w-full px-4 py-2 border border-gray-300 rounded-md bg-white cursor-pointer"
                        />
                    </div>
                </div>

                <button
                    disabled={isUploading}
                    onClick={uploadSong}
                    className="w-full bg-indigo-600 text-white font-semibold py-3 rounded-md mt-6 hover:bg-indigo-700 transition"
                >
                    Upload Song
                </button>
                <h4 className="text-xl mt-8 font-semibold text-indigo-300">Last read tag: {currentTag}</h4>
            </div>
        </>
    )
}
