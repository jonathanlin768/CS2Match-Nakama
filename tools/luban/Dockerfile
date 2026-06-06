FROM mcr.microsoft.com/dotnet/sdk:8.0

# Install tools for downloading and extracting Luban
RUN apt-get update && apt-get install -y p7zip-full wget && rm -rf /var/lib/apt/lists/*

# Download Luban release and extract
ARG LUBAN_VERSION=4.5.0
RUN wget -q "https://github.com/focus-creative-games/luban/releases/download/v${LUBAN_VERSION}/Luban.7z" && \
    7z x Luban.7z -o/tools/ && \
    rm Luban.7z

WORKDIR /workspace
ENTRYPOINT ["dotnet", "/tools/Luban/Luban.dll"]
