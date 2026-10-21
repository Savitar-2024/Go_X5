package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.

func longestPrefix(words []string) string {
	if len(words) == 0 {
		return ""
	}

	prefix := []rune(words[0])

	for _, word := range words[1:] {
		runes := []rune(word)

		j := 0
		for j < len(prefix) && j < len(runes) {
			if prefix[j] != runes[j] {
				break
			}
			j++
		}

		prefix = prefix[:j]
	}

	return string(prefix)
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	// TODO: найти единственную пару с суммой input.Target.
	seen := make(map[int]int)

	for i, num := range input.Nums {
		complement := input.Target - num

		if j, exists := seen[complement]; exists {
			return [2]int{j, i}
		}

		seen[num] = i
	}

	return [2]int{}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	// TODO: проверить наличие повторов.
	seen := make(map[int]bool)

	for _, num := range nums {
		if seen[num] {
			return true
		}

		seen[num] = true
	}

	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	// TODO: сохранить порядок и не изменять исходный слайс.
	if len(nums) == 0 {
		return []int{}
	}

	result := make([]int, 0, len(nums))
	result = append(result, nums[0])

	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[i-1] {
			result = append(result, nums[i])
		}
	}

	return result
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	// TODO: обработать строку из символов ()[]{}.
	stack := make([]rune, 0)

	for _, ch := range text {
		switch ch {
		case '(', '[', '{':
			stack = append(stack, ch)

		case ')', ']', '}':
			if len(stack) == 0 {
				return false
			}

			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if ch == ')' && last != '(' ||
				ch == ']' && last != '[' ||
				ch == '}' && last != '{' {
				return false
			}
		}
	}

	return len(stack) == 0
}
