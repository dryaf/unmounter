package main

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Unmounter</title>
	<link href="https://cdn.jsdelivr.net/npm/@picocss/pico@2.0.6/css/pico.min.css" rel="stylesheet"/>
	<link href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.colors.min.css" rel="stylesheet"/>
	<style>
		main {
			max-width: 1024px; 
			margin: auto;
		}
		.bi {
			margin-right: 5px;
		}
		.mount-group {
			margin:0;
			padding:0;
		}
		.multiline {
			white-space: pre-wrap;
		}
		.service {
			font-size: 1.2em;
			display: inline-block; 
			width: 200px;
			vertical-align: middle; 
			margin-right: 10px; 
			margin-top: 0.6em;
		}
		.usb-icon {
			display: inline-block; 
			vertical-align: middle; 
			margin-right: 10px; 
			margin-top: 0.6em;
		}
		.clean {
			margin: 0;
			padding: 0;
		}
		input[type="submit"], button {
			padding: 0.5em 1em;
			vertical-align: middle;
		}
  	</style>
	</head>
	<body>
	<main class="container">
		<nav>
			<ul>
				<li><h1><svg xmlns="http://www.w3.org/2000/svg" height="0.8em" fill="currentColor" class="bi bi-tools" viewBox="0 0 16 16">
				<path d="M1 0 0 1l2.2 3.081a1 1 0 0 0 .815.419h.07a1 1 0 0 1 .708.293l2.675 2.675-2.617 2.654A3.003 3.003 0 0 0 0 13a3 3 0 1 0 5.878-.851l2.654-2.617.968.968-.305.914a1 1 0 0 0 .242 1.023l3.27 3.27a.997.997 0 0 0 1.414 0l1.586-1.586a.997.997 0 0 0 0-1.414l-3.27-3.27a1 1 0 0 0-1.023-.242L10.5 9.5l-.96-.96 2.68-2.643A3.005 3.005 0 0 0 16 3q0-.405-.102-.777l-2.14 2.141L12 4l-.364-1.757L13.777.102a3 3 0 0 0-3.675 3.68L7.462 6.46 4.793 3.793a1 1 0 0 1-.293-.707v-.071a1 1 0 0 0-.419-.814zm9.646 10.646a.5.5 0 0 1 .708 0l2.914 2.915a.5.5 0 0 1-.707.707l-2.915-2.914a.5.5 0 0 1 0-.708M3 11l.471.242.529.026.287.445.445.287.026.529L5 13l-.242.471-.026.529-.445.287-.287.445-.529.026L3 15l-.471-.242L2 14.732l-.287-.445L1.268 14l-.026-.529L1 13l.242-.471.026-.529.445-.287.287-.445.529-.026z"/>
				  </svg> 
				  Unmounter</h1></li>
			</ul>
			<ul>
				<li><button href="#" onclick="this.disabled=true; setTimeout(function(){ location.reload(); }, 500);" role="button" class="primary">Refresh</button></li>
				<li><a class="pico-color-slate-850" href="https://github.com/dryaf/unmounter"><svg xmlns="http://www.w3.org/2000/svg" width="1.8em" fill="currentColor" class="bi bi-github" viewBox="0 0 16 16">
				<path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8"/>
			  </svg></a></li>
			</ul>
		</nav>
		<section>
			<h2>Services</h2>
			<details class="outline">
				<summary>

					<div class="service">
						<svg xmlns="http://www.w3.org/2000/svg" width="0.8em" fill="currentColor" class="bi bi-info-circle" viewBox="0 0 16 16">
						<path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/>
						<path d="m8.93 6.588-2.29.287-.082.38.45.083c.294.07.352.176.288.469l-.738 3.468c-.194.897.105 1.319.808 1.319.545 0 1.178-.252 1.465-.598l.088-.416c-.2.176-.492.246-.686.246-.275 0-.375-.193-.304-.533zM9 4.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0"/>
						</svg>
						Autofs
					</div> 
					{{with .Autofs.Active}}<kbd class="pico-background-green">active</kbd>{{else}}<kbd  class="pico-background-red">inactive !!!</kbd>{{end}}

				</summary>
				<kbd class="multiline overflow-auto">
					{{.Autofs.Detail}}
				</kbd>
			</details>
			<details  class="outline">
				<summary>

					<div class="service">
						<svg xmlns="http://www.w3.org/2000/svg" width="0.8em" fill="currentColor" class="bi bi-info-circle" viewBox="0 0 16 16">
						<path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14m0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16"/>
						<path d="m8.93 6.588-2.29.287-.082.38.45.083c.294.07.352.176.288.469l-.738 3.468c-.194.897.105 1.319.808 1.319.545 0 1.178-.252 1.465-.598l.088-.416c-.2.176-.492.246-.686.246-.275 0-.375-.193-.304-.533zM9 4.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0"/>
						</svg>
						Samba
					</div> 
					{{with .Samba.Active}} <kbd class="pico-background-green">no locked files</kbd>{{else}}<kbd  class="pico-background-red">LOCKED FILES!!!</kbd>{{end}}

				</summary>
				<kbd class="multiline overflow-auto">
					{{.Samba.Detail}}
				</kbd>
			</details>
		</section>
		<hr />
		<section>	
			<h2>Mounted Devices</h1>

			<form action="/restart-autofs" method="post">
				<input type="submit" role="button" class="primary" value="Restart Autofs" />
			</form>

			{{ if .Error }}
				<mark class="pico-background-red">{{ .Error }}</mark>
			{{ end }}

			{{ if not .Mounts }}
				<p>No devices mounted at /mnt or /media.</p>
			{{ else }}
				{{range $i, $m :=  .Mounts }}
				<article>
					<header class="clean">
						<form  action="/" method="post" style="margin: 0;">
							<input name="device" type="hidden" value="{{$m.Path}}"/>
							<fieldset role="group" class="clean">
								<span  class="usb-icon" title="{{$m.Device}}"><svg xmlns="http://www.w3.org/2000/svg" width="1.8em" fill="currentColor" class="bi bi-usb-drive" viewBox="0 0 16 16">
								<path d="M6 .5a.5.5 0 0 1 .5-.5h4a.5.5 0 0 1 .5.5v4H6zM7 1v1h1V1zm2 0v1h1V1zM6 5a1 1 0 0 0-1 1v8.5A1.5 1.5 0 0 0 6.5 16h4a1.5 1.5 0 0 0 1.5-1.5V6a1 1 0 0 0-1-1zm0 1h5v8.5a.5.5 0 0 1-.5.5h-4a.5.5 0 0 1-.5-.5z"/>
							</svg></span>
								<input type="text" title="{{$m.Device}}" value="{{ $m.Path }}" disabled />
								{{with $m.Usages}}
									<span  data-tooltip="cannot unmount because it is in use"><input type="submit" value="Unmount" disabled/></span>
								{{else}}
									<input type="submit" value="Unmount"/>
								{{end}}
							</fieldset>
						</form>
					</header>
					{{with $m.UsageError}}
						<mark class="pico-background-red">Error fetching usages: {{.}}</mark>
					{{end}}
					{{with $m.Usages}}
						<table>
							<thead>
								<tr>
									<th scope="col" class="pico-color-purple-600">in use by</th>
									<th scope="col">PID</th>
									<th scope="col">USER</th>
									<th scope="col">NAME</th>
								</tr>
							</thead>
							<tbody>
								{{ range . }}
									<tr>
										<th scope="row">{{.Command}}</th>
										<td>{{.PID}}</td>
										<td>{{.User}}</td>
										<td>{{.Name}}</td>
									</tr>
								{{end}}
							</tbody>
						</table>
					{{end}}
				</article>
				{{ end }}
			{{ end }}
		</section>

	</main>
	<footer class="container">
		
	</footer>
</body>
</html>
`
