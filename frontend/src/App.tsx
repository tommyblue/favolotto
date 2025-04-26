import React from "react"
import List from "./components/List"
import Head from "./components/Head"
import Form from "./components/Form"
import "./App.scss"

function App() {
	return (
		<div className="bg-gray-800 min-h-screen flex flex-col items-center justify-center px-6 sm:px-8 lg:px-12 py-10 bg-gray-100">
			<Head />
			<Form />
			<List />
		</div>
	)
}

export default App
