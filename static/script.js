import { get, set } from "https://cdn.jsdelivr.net/npm/idb-keyval@6/+esm";

const prefixMap = new Map();
const upstreamMap = new Map();
const tbody = document.getElementById("tbody");
let asnr = 200020;

const asnamecb = {};

if (window.location.hash) {
	asnr = parseInt(window.location.hash.substring(1));
}

function processEvent(evt) {
	// const d = moment.unix(evt.Timestamp);
	const d = moment(evt.Timestamp);
	const dstAS = evt.DestinationAS;
	const ispAS = evt.ISPAS;
	const transitAS = evt.TransitAS;
	const prefix = evt.Prefix;

	if (!prefixMap.has(prefix)) {
		prefixMap.set(prefix, {
			dstAS: dstAS,
			ispAS: new Set(),
			transitAS: new Set(),
			PrefixWhois: evt.PrefixWhois,
			timestamps: []
		});
		upstreamMap.set(prefix, new Set());
	}

	prefixMap.get(prefix).timestamps.push(d);
	prefixMap.get(prefix).ispAS.add(ispAS);
	prefixMap.get(prefix).transitAS.add(transitAS);
	upstreamMap.get(prefix).add(ispAS);

	return prefix;
}

function createRow(prefix, value) {
	let row = document.getElementById(`prefix-${prefix}`);
	if (row) {
		tbody.removeChild(row);
		console.log("updating row", prefix);
	}
	row = document.createElement("tr");
	row.id = `prefix-${prefix}`;

	const now = moment();

	let avgtime = value.timestamps.map((t) => now.diff(t)).reduce((s, c) => s + c, 0) / value.timestamps.length;
	row.style.opacity = Math.min(1, Math.max(0.25, 1 - ((avgtime - (3000 * 3600)) / (3600 * 1000 * 12))));
	let d = 1 - (avgtime / (1000 * 3600 * 2));
	let medtime = value.timestamps[Math.floor(value.timestamps.length / 2)];

	let firstDateTD = document.createElement("td");
	if (d > 0) {
		firstDateTD.style.background = 'rgba(255,0,0,' + d / 2 + ')';
	}

	row.appendChild(firstDateTD);
	firstDateTD.textContent = value.timestamps[0].fromNow();

	let lastDateTD = document.createElement("td");
	row.appendChild(lastDateTD);
	lastDateTD.textContent = value.timestamps.at(-1).fromNow();

	let countTD = document.createElement("td");
	row.appendChild(countTD);
	countTD.textContent = value.timestamps.length + "/" + upstreamMap.get(prefix).size;

	let whoisTD = document.createElement("td");
	row.appendChild(whoisTD);
	whoisTD.textContent = value.PrefixWhois;

	let dstASTD = document.createElement("td");
	row.appendChild(dstASTD);
	dstASTD.textContent = `AS${value.dstAS.Num}`;

	let dstASNameTD = document.createElement("td");
	row.appendChild(dstASNameTD);
	dstASNameTD.textContent = `AS${value.dstAS.Name}`;

	let prefixTD = document.createElement("td");
	row.appendChild(prefixTD);
	prefixTD.textContent = prefix;

	let dnslink = document.createElement("a");
	//dnslink.href = `https://bgp.he.net/net/${prefix}#_dnsrecords`;
	dnslink.href = `https://bgp.he.net/search?search[search]=${prefix}&commit=Search`;
	dnslink.textContent = "👁️";
	prefixTD.appendChild(dnslink);

	tbody.insertBefore(row, tbody.firstChild);
}

function repaint() {
	let elems = [...prefixMap.entries()];
	elems.sort((a, b) => {
		return a[1].timestamps.at(-1) - b[1].timestamps.at(-1);
	});

	elems.forEach((e) => {
		createRow(e[0], e[1]);
	});
}


var url = new URL(`/api/ws?asnr=${asnr}`, window.location.href);
url.protocol = url.protocol.replace('http', 'ws');

var ws = new WebSocket(url.href);
ws.onmessage = function (event) {
	let msg = JSON.parse(event.data);
	let d = processEvent(msg);
	document.title = msg.TransitAS.Name;
	createRow(d, prefixMap.get(d));
};
