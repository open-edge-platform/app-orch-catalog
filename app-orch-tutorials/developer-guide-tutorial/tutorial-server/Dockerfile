# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# Use the FastAPI image from Docker Hub
FROM demisto/fastapi:0.115.9.2587379@sha256:367f6b9a2179122edde70e3a9b5a3e88db7b9ee9de16dcd00a99d06c92987a7a

# Set the working directory
WORKDIR /app

# Copy the requirements file into the container
COPY main.py ./

# Expose the port the app runs on
EXPOSE 8000

# Run as a non-root user
USER nobody

# Run the application
ENTRYPOINT ["uvicorn", "main:app", "--host", "0.0.0.0"]

