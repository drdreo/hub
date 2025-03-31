using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Hosting;

var app = WebApplication.Create();
app.MapGet("/", () => "Hello, Tech Demo!");
app.Urls.Add("http://*:8080");
app.Run();
