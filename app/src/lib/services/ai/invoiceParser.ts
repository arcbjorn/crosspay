export interface ExtractedField {
	name: string;
	value: string;
	confidence: number; // 0 to 1
	coordinates?: {
		x: number;
		y: number;
		width: number;
		height: number;
	};
}

export interface InvoiceData {
	invoiceNumber?: string;
	date?: string;
	dueDate?: string;
	vendor?: {
		name: string;
		address?: string;
		email?: string;
		phone?: string;
	};
	billTo?: {
		name: string;
		address?: string;
	};
	items: Array<{
		description: string;
		quantity: number;
		unitPrice: number;
		total: number;
	}>;
	subtotal?: number;
	tax?: number;
	total: number;
	currency?: string;
}

export interface ParsingResult {
	success: boolean;
	data: InvoiceData | null;
	extractedFields: ExtractedField[];
	confidence: number;
	processingTime: number;
	errors: string[];
}

class InvoiceParserAI {
	private commonPatterns = {
		invoiceNumber: [
			/invoice[#\s]*:?\s*([A-Z0-9\-]+)/i,
			/inv[#\s]*:?\s*([A-Z0-9\-]+)/i,
			/reference[#\s]*:?\s*([A-Z0-9\-]+)/i
		],
		date: [
			/date[:\s]+(\d{1,2}[\/\-]\d{1,2}[\/\-]\d{2,4})/i,
			/(\d{1,2}[\/\-]\d{1,2}[\/\-]\d{2,4})/,
			/(\d{4}-\d{2}-\d{2})/
		],
		amount: [
			/total[:\s]*\$?([0-9,]+\.?\d{0,2})/i,
			/amount[:\s]*\$?([0-9,]+\.?\d{0,2})/i,
			/\$([0-9,]+\.?\d{0,2})/
		],
		vendor: [/from[:\s]+([A-Za-z\s&.,'-]+)(?:\n|$)/i, /vendor[:\s]+([A-Za-z\s&.,'-]+)(?:\n|$)/i]
	};

	async parseInvoice(file: File): Promise<ParsingResult> {
		const startTime = Date.now();
		const errors: string[] = [];

		try {
			// Simulate OCR text extraction
			const extractedText = await this.extractTextFromFile(file);

			if (!extractedText) {
				return {
					success: false,
					data: null,
					extractedFields: [],
					confidence: 0,
					processingTime: Date.now() - startTime,
					errors: ['Failed to extract text from file']
				};
			}

			// Parse the extracted text
			const extractedFields = this.extractFields(extractedText);
			const invoiceData = this.buildInvoiceData(extractedFields);
			const confidence = this.calculateConfidence(extractedFields);

			return {
				success: confidence > 0.5,
				data: invoiceData,
				extractedFields,
				confidence,
				processingTime: Date.now() - startTime,
				errors
			};
		} catch (error) {
			errors.push(`Parsing error: ${error instanceof Error ? error.message : 'Unknown error'}`);

			return {
				success: false,
				data: null,
				extractedFields: [],
				confidence: 0,
				processingTime: Date.now() - startTime,
				errors
			};
		}
	}

	private async extractTextFromFile(file: File): Promise<string> {
		// This is a simplified mock implementation
		// In a real implementation, you would use an OCR library like Tesseract.js
		// or call a cloud OCR service

		return new Promise((resolve) => {
			// Simulate processing time
			setTimeout(
				() => {
					// Mock extracted text based on file name or type
					if (file.name.includes('invoice')) {
						resolve(`
            INVOICE #INV-2024-001
            Date: 03/15/2024
            Due Date: 04/15/2024
            
            From: TechCorp Solutions
            123 Business Ave
            San Francisco, CA 94107
            
            Bill To: CrossPay Protocol
            456 Innovation St
            New York, NY 10001
            
            Description                 Qty    Rate      Total
            Software License            1      $1,000    $1,000
            Support Services           12      $200      $2,400
            Implementation             1       $5,000    $5,000
            
            Subtotal:                             $8,400
            Tax (8.5%):                           $714
            TOTAL:                                $9,114
          `);
					} else {
						resolve('');
					}
				},
				1000 + Math.random() * 2000
			); // 1-3 seconds
		});
	}

	private extractFields(text: string): ExtractedField[] {
		const fields: ExtractedField[] = [];

		// Extract invoice number
		for (const pattern of this.commonPatterns.invoiceNumber) {
			const match = text.match(pattern);
			if (match) {
				fields.push({
					name: 'invoiceNumber',
					value: match[1],
					confidence: 0.9
				});
				break;
			}
		}

		// Extract dates
		for (const pattern of this.commonPatterns.date) {
			const matches = [...text.matchAll(new RegExp(pattern.source, pattern.flags + 'g'))];
			matches.forEach((match, index) => {
				fields.push({
					name: index === 0 ? 'date' : 'dueDate',
					value: match[1],
					confidence: 0.8
				});
			});
		}

		// Extract amounts
		const amounts = [...text.matchAll(/\$([0-9,]+\.?\d{0,2})/g)];
		amounts.forEach((match, index) => {
			const value = match[1].replace(/,/g, '');
			const fieldName =
				index === amounts.length - 1
					? 'total'
					: index === amounts.length - 2
						? 'tax'
						: index === amounts.length - 3
							? 'subtotal'
							: `amount_${index}`;

			fields.push({
				name: fieldName,
				value,
				confidence: 0.85
			});
		});

		// Extract vendor info
		const vendorMatch = text.match(/From:\s*([^\n]+)/i);
		if (vendorMatch) {
			fields.push({
				name: 'vendorName',
				value: vendorMatch[1].trim(),
				confidence: 0.9
			});
		}

		// Extract line items
		const lineItemPattern = /(.+?)\s+(\d+)\s+\$([0-9,]+\.?\d{0,2})\s+\$([0-9,]+\.?\d{0,2})/g;
		const lineItems = [...text.matchAll(lineItemPattern)];

		lineItems.forEach((match, index) => {
			fields.push({
				name: `item_${index}_description`,
				value: match[1].trim(),
				confidence: 0.8
			});
			fields.push({
				name: `item_${index}_quantity`,
				value: match[2],
				confidence: 0.9
			});
			fields.push({
				name: `item_${index}_rate`,
				value: match[3].replace(/,/g, ''),
				confidence: 0.9
			});
			fields.push({
				name: `item_${index}_total`,
				value: match[4].replace(/,/g, ''),
				confidence: 0.9
			});
		});

		return fields;
	}

	private buildInvoiceData(fields: ExtractedField[]): InvoiceData {
		const getFieldValue = (name: string) => fields.find((f) => f.name === name)?.value;

		// Extract line items
		const items: InvoiceData['items'] = [];
		let itemIndex = 0;

		while (getFieldValue(`item_${itemIndex}_description`)) {
			const description = getFieldValue(`item_${itemIndex}_description`) || '';
			const quantity = parseFloat(getFieldValue(`item_${itemIndex}_quantity`) || '0');
			const unitPrice = parseFloat(getFieldValue(`item_${itemIndex}_rate`) || '0');
			const total = parseFloat(getFieldValue(`item_${itemIndex}_total`) || '0');

			items.push({
				description,
				quantity,
				unitPrice,
				total
			});

			itemIndex++;
		}

		return {
			invoiceNumber: getFieldValue('invoiceNumber'),
			date: getFieldValue('date'),
			dueDate: getFieldValue('dueDate'),
			vendor: {
				name: getFieldValue('vendorName') || ''
			},
			items,
			subtotal: parseFloat(getFieldValue('subtotal') || '0') || undefined,
			tax: parseFloat(getFieldValue('tax') || '0') || undefined,
			total: parseFloat(getFieldValue('total') || '0'),
			currency: 'USD'
		};
	}

	private calculateConfidence(fields: ExtractedField[]): number {
		if (fields.length === 0) return 0;

		// Weight different field types
		const weights = {
			invoiceNumber: 0.2,
			date: 0.15,
			total: 0.3,
			vendorName: 0.15,
			subtotal: 0.1,
			tax: 0.1
		};

		let totalWeight = 0;
		let weightedConfidence = 0;

		for (const field of fields) {
			const weight = weights[field.name as keyof typeof weights] || 0.05;
			totalWeight += weight;
			weightedConfidence += field.confidence * weight;
		}

		// Bonus for having required fields
		const hasInvoiceNumber = fields.some((f) => f.name === 'invoiceNumber');
		const hasTotal = fields.some((f) => f.name === 'total');
		const hasVendor = fields.some((f) => f.name === 'vendorName');

		let bonus = 0;
		if (hasInvoiceNumber && hasTotal && hasVendor) {
			bonus = 0.1; // 10% bonus for having all key fields
		}

		return Math.min(1, weightedConfidence / totalWeight + bonus);
	}

	// Utility method for terminal-style progress display
	generateParsingProgress(step: number, total: number): string {
		const progress = Math.floor((step / total) * 20);
		const bar = '█'.repeat(progress) + '░'.repeat(20 - progress);
		const percentage = Math.floor((step / total) * 100);

		return `[${bar}] ${percentage}%`;
	}

	// Generate terminal-style field extraction display
	generateFieldExtractionDisplay(fields: ExtractedField[]): string {
		let output = 'FIELD_EXTRACTION_RESULTS:\n';
		output += '┌─────────────────────────────────────────────────────────┐\n';

		fields.forEach((field) => {
			const confidence = Math.floor(field.confidence * 100);
			const confidenceBar =
				'█'.repeat(Math.floor(confidence / 10)) + '░'.repeat(10 - Math.floor(confidence / 10));

			output += `│ ${field.name.toUpperCase().padEnd(20)} │ ${field.value.padEnd(15)} │ [${confidenceBar}] ${confidence}% │\n`;
		});

		output += '└─────────────────────────────────────────────────────────┘';
		return output;
	}
}

export const invoiceParserAI = new InvoiceParserAI();
