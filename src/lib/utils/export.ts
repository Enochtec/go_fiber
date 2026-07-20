function esc(s: string): string {
	const v = String(s ?? '');
	return v.includes(',') || v.includes('"') || v.includes('\n') ? `"${v.replace(/"/g, '""')}"` : v;
}

export function exportCsv(filename: string, headers: string[], rows: string[][], title = '') {
	const now = new Date();
	const dateStr = now.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
	const timeStr = now.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });

	const meta = title ? `${esc(title)}\n` : '';
	const header = `Exported,${esc(dateStr)}\nTime,${esc(timeStr)}\n\n`;
	const colLine = headers.join(',');
	const dataLines = rows.map(r => r.map(esc).join(',')).join('\n');

	const csv = meta + header + colLine + '\n' + dataLines + '\n';

	const bom = '\uFEFF';
	const blob = new Blob([bom + csv], { type: 'text/csv;charset=utf-8;' });
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = `${filename}-${now.toISOString().slice(0, 10)}.csv`;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}
