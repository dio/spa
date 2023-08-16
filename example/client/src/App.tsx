import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'

import { Routes, Route, Link } from "react-router-dom";

function Index() {
  return (
    <>
      <Logos to="/next"/>
      <h1>Vite + React</h1>
      <p className="read-the-docs">
        Click on the Vite and React logos to go to the <Link to="/next">next</Link> page
      </p>
    </>
  )
}

function Next() {
  return (
    <>
      <Logos to="/"/>
      <h1>Next</h1>
      <p className="read-the-docs">
        Click on the Vite and React logos to go <Link to="/">back</Link>
      </p>
    </>
  )
}

function Logos(props: any) {
  return(
    <>
    <div>
      <Link to={props.to}>
        <img src={viteLogo} className="logo" alt="Vite logo" />
      </Link>
      <Link to={props.to}>
        <img src={reactLogo} className="logo react" alt="React logo" />
      </Link>
    </div>
    </>
  )
}

function App() {
  return (
    <Routes>
      <Route path='/' element={<Index />} />
      <Route path='/next' element={<Next />} />
    </Routes>
  )
}

export default App
