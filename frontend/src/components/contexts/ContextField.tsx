import { Input } from "../ui/input.js";

interface ContextFieldProps {
	label: string;
	value: string;
	onChange: (value: string) => void;
}

export function ContextField({ label, value, onChange }: ContextFieldProps) {
	return (
		<label className="block text-sm">
			{label}
			<Input
				className="mt-1"
				value={value}
				onChange={(event) => onChange(event.target.value)}
			/>
		</label>
	);
}
