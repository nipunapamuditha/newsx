# NewsX

NewsX is a web service that aggregates content from Substack newsletters, generates news scripts using Azure AI, and converts them into audio files for easy consumption.

## Demo

Try the live demo: [https://demo.newsloop.xyz](https://demo.newsloop.xyz)

API endpoint: [https://newsxapi.newsloop.xyz](https://newsxapi.newsloop.xyz)

## Features

- **User Authentication**: Secure login with Google OAuth
- **Content Aggregation**: Fetch latest articles from Substack newsletters
- **AI Script Generation**: Convert articles into a coherent news script using Azure's language models
- **Text-to-Speech**: Generate professional-quality audio files from the scripts
- **User Preferences**: Save and manage your preferred content sources
- **Audio Management**: Access and delete generated audio files

## Technology Stack

- **Backend**: Go (Gin framework)
- **Database**: MySQL
- **Authentication**: JWT, Google OAuth
- **Storage**: MinIO object storage
- **AI Services**: Azure Cognitive Services (Language Models, Speech Synthesis)
- **CI/CD**: Jenkins

## API Endpoints

### Authentication
- `POST /v1/signuporlogin`: Sign up or log in with Google OAuth token

### User Management
- `GET /v1/getuser`: Get current user information
- `POST /v1/update-preferences`: Update user preferences
- `POST /v1/publish-preferances`: Save user content source preferences
- `GET /v1/Get_Preferances`: Retrieve user preferences

### Content Generation
- `GET /v1/Generate_now`: Generate news audio (server-sent events for progress updates)
- `GET /v1/getaudio_files`: Get list of user's generated audio files
- `POST /v1/delete_audiofile`: Delete a specific audio file

### Twitter Integration
- `POST /v1/getTwUsernames`: Retrieve Twitter user information

## Setup and Installation

### Prerequisites
- Go 1.23.4 or later
- MySQL database
- MinIO server
- Azure account with Cognitive Services enabled

### Environment Variables
Create a `.env` file with the following variables:

- DB_HOST=127.0.0.1 
- DB_PORT=3306 DB_USER=your_db_user 
- DB_PASSWORD=your_db_password
- DB_NAME=newx 
- WT_SECRET=your_jwt_secret_key 
- GOOGLE_CLIENT_ID=your_google_client_id 
- bearer_tk_for_user_search=twitter_bearer_token 
- Azure_LLM_Key=your_azure_llm_key 
- MINIO_ACCESS_KEY=your_minio_access_key 
- MINIO_SECRET_KEY=your_minio_secret_key 
- AZURE_SPEECH_KEY=your_azure_speech_key

### Build and Run
```bash
# Build the application
go build -o app

# Run the application
./app
