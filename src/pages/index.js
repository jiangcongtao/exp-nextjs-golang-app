// pages/index.js
import React, { useState } from 'react';

export default function Home() {
  const [inputText, setInputText] = useState('');
  const [responseText, setResponseText] = useState('');

  async function handleSubmit() {
    const response = await fetch('/process', {
      method: 'POST',
      body: inputText,
      headers: {
        'Content-Type': 'text/plain',
      },
    });
    const text = await response.text();
    setResponseText(text);
  }

  return (
    <div>
      <h1>Next.js & Golang App</h1>
      <textarea
        value={inputText}
        onChange={(e) => setInputText(e.target.value)}
        style={{ width: '100%', minHeight: '100px' }}
      />
      <button onClick={handleSubmit} style={{ display: 'block', marginTop: '10px' }}>
        Send to Server
      </button>
      {responseText && (
        <div style={{ backgroundColor: 'blue', color: 'white', marginTop: '10px', padding: '10px' }}>
          {responseText}
        </div>
      )}
    </div>
  );
}
