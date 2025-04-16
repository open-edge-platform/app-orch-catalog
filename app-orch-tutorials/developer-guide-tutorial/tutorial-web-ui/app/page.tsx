'use client'

import {useEffect, useState} from 'react';
import axios from 'axios';

// Axios Interceptor Instance
const AxiosInstance = axios.create({
  baseURL: process.env.NODE_ENV === 'development' ? 'http://localhost:8000' : '/api'
});

export default function Home() {
  const [count, setCount] = useState(0);
  const [greeting, setGreeting] = useState("not yet set");
  const [error, setError] = useState(null);

  useEffect(() => {
    AxiosInstance.get('/counter')
        .then(response => {
          setCount(response.data.count);
        })
        .catch(error => {
          setError(error.message);
        });
  }, []);

  useEffect(() => {
    AxiosInstance.get('/')
        .then(response => {
          setGreeting(response.data.message);
        })
        .catch(error => {
          setError(error.message);
        });
  }, []);

  return (
      <div
          className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
        <header className="row-start-1 flex gap-[24px] flex-wrap items-center justify-center">
          <div className="flex items-center bg-blue-500 text-white text-sm font-bold px-4 py-3" role="alert">
            <svg className="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
              <path
                  d="M12.432 0c1.34 0 2.01.912 2.01 1.957 0 1.305-1.164 2.512-2.679 2.512-1.269 0-2.009-.75-1.974-1.99C9.789 1.436 10.67 0 12.432 0zM8.309 20c-1.058 0-1.833-.652-1.093-3.524l1.214-5.092c.211-.814.246-1.141 0-1.141-.317 0-1.689.562-2.502 1.117l-.528-.88c2.572-2.186 5.531-3.467 6.801-3.467 1.057 0 1.233 1.273.705 3.23l-1.391 5.352c-.246.945-.141 1.271.106 1.271.317 0 1.357-.392 2.379-1.207l.6.814C12.098 19.02 9.365 20 8.309 20z"/>
            </svg>
            <p>{greeting}</p>
          </div>
        </header>
        <main className="flex flex-col gap-[32px] row-start-2 items-center sm:items-start">
          <div className="bg-blue-100 border-t border-b border-blue-500 text-blue-700 px-4 py-3" role="alert">
            <p className="font-bold">Counter</p>
            <p className="text-sm">{count}</p>
          </div>
          <div>
            <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-full">
              <input type="button" value="Increment" onClick={() => {
                AxiosInstance.post('/increment')
                    .then(response => {
                      setCount(response.data.count);
                    })
                    .catch(error => {
                      setError(error.message);
                    });
              }}/>
            </button>
            <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-full">
              <input type="button" value="Decrement" onClick={() => {
                AxiosInstance.post('/decrement')
                    .then(response => {
                      setCount(response.data.count);
                    })
                    .catch(error => {
                      setError(error.message);
                    });
              }}/>
            </button>
            <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-full">
              <input type="button" value="Reinitialize" onClick={() => {
                AxiosInstance.post('/reinitialize')
                    .then(response => {
                      setCount(response.data.count);
                    })
                    .catch(error => {
                      setError(error.message);
                    });
              }}/>
            </button>
          </div>
        </main>
        <footer className="row-start-3 flex gap-[24px] flex-wrap items-center justify-center">
          {error && <p>Error: {error}</p>}
        </footer>
      </div>
  );
}