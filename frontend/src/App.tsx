import {useState} from 'react';
import logo from './assets/images/logo-universal.png';
import {devContextApi} from "./lib/devctx-api";

function App() {
    const [resultText, setResultText] = useState("Please enter your name below 👇");
    const [name, setName] = useState('');
    const updateName = (e: any) => setName(e.target.value);
    const updateResultText = (result: string) => setResultText(result);

    function greet() {
        devContextApi.greet(name).then((result) => {
            if (result.ok) {
                updateResultText(result.data);
                return;
            }
            updateResultText(result.error.message);
        });
    }

    return (
        <div id="App">
            <img src={logo} id="logo" alt="logo"/>
            <div id="result" className="result">{resultText}</div>
            <div id="input" className="input-box">
                <input id="name" className="input" onChange={updateName} autoComplete="off" name="input" type="text"/>
                <button className="btn" onClick={greet}>Greet</button>
            </div>
        </div>
    )
}

export default App
