interface ContextFieldProps {
	label: string;
	value: string;
	onChange: (value: string) => void;
}

export function ContextField({ label, value, onChange }: ContextFieldProps) {
	return (
		<label className="block text-sm">
			{label}
			<input
				className="mt-1 w-full border p-2"
				value={value}
				onChange={(event) => onChange(event.target.value)}
			/>
		</label>
	);
}
