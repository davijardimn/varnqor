import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './AppV3'
import './v3.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode><App /></StrictMode>,
)
