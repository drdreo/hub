FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /app

COPY apps/csharp-api/ ./
RUN dotnet restore
RUN dotnet publish -o out

FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app
COPY --from=build /app/out .
EXPOSE 8080
ENTRYPOINT ["dotnet", "csharp-api.dll"]
