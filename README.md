# Substack Copilot (NewsX v3)

This project is a Go-based backend service that fetches articles from specified Substack feeds, generates a summarized news script using Azure AI, synthesizes the script into an audio file using Azure Text-to-Speech, and stores the audio file in a MinIO bucket. It provides a RESTful API for user management, preference handling, and audio generation/retrieval.

## Features

*   **User Authentication:** Sign up and login using Google OAuth. JWT-based session management.
*   **Substack Integration:** Fetches recent articles from user-specified Substack RSS feeds.
*   **AI Script Generation:** Uses Azure AI (DeepSeek model) to generate an engaging news script summarizing the fetched articles.
*   **Text-to-Speech:** Converts the generated script into a WAV audio file using Azure Cognitive Services Speech SDK.
*   **Audio Storage:** Stores generated audio files in a MinIO S3-compatible object storage bucket, organized by user.
*   **Preference Management:** Allows users to save and retrieve their preferred Substack usernames.
*   **Streaming Generation:** Provides real-time updates during the audio generation process via Server-Sent Events (SSE).
*   **API:** Exposes endpoints for managing users, preferences, generating audio, and accessing/deleting generated files.
*   **CI/CD:** Includes a [Jenkinsfile](e:\Personal Projects\newsx_version_3\Jenkinsfile) for automated build and deployment.

## Tech Stack

*   **Backend:** Go (Golang)
*   **Web Framework:** Gin
*   **Database:** MySQL
*   **Authentication:** JWT, Google OAuth
*   **AI Services:** Azure AI (LLM), Azure Cognitive Services (TTS)
*   **Object Storage:** MinIO
*   **CI/CD:** Jenkins
*   **Environment Variables:** godotenv

## Setup

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd newsx_version_3
    ```
2.  **Create a `.env` file:**
    Copy the `.env.example` (if available) or create a new `.env` file in the root directory and populate it with the necessary environment variables:
    *   `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (for MySQL)
    *   `GOOGLE_CLIENT_ID` (for Google OAuth)
    *   `JWT_SECRET` (for JWT signing)
    *   `AZURE_LLM_KEY` (for Azure AI script generation)
    *   `AZURE_SPEECH_KEY` (for Azure TTS)
    *   `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY` (for MinIO storage)
3.  **Install Go dependencies:**
    ```bash
    go mod tidy
    ```
4.  **Build the application:**
    ```bash
    go build -o app
    ```
5.  **Run the application:**
    ```bash
    ./app
    ```
    The API will be available at `http://localhost:9090`.

## API Endpoints (v1)

All endpoints are prefixed with `/v1`. Authentication is required for most endpoints via a JWT cookie set during login.

*   `POST /signuporlogin`: Handles user signup or login using a Google Auth token in the request body. Sets an `Authorization` cookie on success.
*   `POST /getTwUsernames`: (Requires Auth) Validates a Twitter username (placeholder functionality).
*   `GET /getuser`: (Requires Auth) Returns the authenticated user's details.
*   `POST /update-preferences`: (Requires Auth) Updates user's Twitter preferences (placeholder functionality).
*   `POST /publish-preferances`: (Requires Auth) Saves the list of Substack usernames provided in the request body to the user's preferences.
*   `GET /Get_Preferances`: (Requires Auth) Retrieves the authenticated user's saved Substack preferences.
*   `GET /Generate_now`: (Requires Auth) Starts the Substack fetch, script generation, and audio synthesis process. Streams progress updates via SSE.
*   `GET /getaudio_files`: (Requires Auth) Returns a list of presigned URLs for the user's generated audio files stored in MinIO.
*   `POST /delete_audiofile`: (Requires Auth) Deletes a specific audio file from MinIO based on the object name provided in the request body.

## Deployment

*   A [Jenkinsfile](e:\Personal Projects\newsx_version_3\Jenkinsfile) is provided for automating the build and deployment process.
*   The Jenkins pipeline builds the Go application, runs tests, and then uses SSH to deploy the binary (`app`) and the [restart-service.sh](e:\Personal Projects\newsx_version_3\restart-service.sh) script to the target server (configured as `10.10.10.41` in the Jenkinsfile).
*   The `restart-service.sh` script handles stopping any existing instance of the application and starting the new version using `nohup`.
*   Logging is configured to write to `newsx.log` ([logging/logging.go](e:\Personal Projects\newsx_version_3\logging\logging.go)).
*   The application binary (`app`) and log file (`newsx.log`) are ignored by Git ([.gitignore](e:\Personal Projects\newsx_version_3\.gitignore)).