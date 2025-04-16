# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

#!/usr/bin/python3

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import os

origins = [
    "http://localhost:3000",
]

app = FastAPI()

"""Necessary to allow CORS (Cross-Origin Resource Sharing) for the web UI"""
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["GET, POST"],
    allow_headers=["*"],
)

class Counter(BaseModel):
    count: int

initial_value = int(os.environ.get('INITIAL_COUNT', '0'))
counter = Counter(count=initial_value)

@app.get("/")
async def read_root():
    """Return a greeting message based on the environment variable TUTORIAL_GREETING"""
    tutorial_greeting = os.environ.get('TUTORIAL_GREETING', 'Hello World')
    return {"message": tutorial_greeting}

@app.get("/counter")
async def read_counter():
    """Return the current count"""
    return counter

@app.post("/increment")
async def increment_counter():
    """Increase the counter by 1 and return it"""
    counter.count += 1
    return counter

@app.post("/decrement")
async def decrement_counter():
    """Decrease the counter by 1 and return it"""
    counter.count -= 1
    return counter

@app.post("/reinitialize")
async def reinitialize_counter():
    """Reinitialize the counter to the initial value and return it"""
    counter.count = initial_value
    return counter