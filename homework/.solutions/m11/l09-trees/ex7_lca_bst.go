package l09trees

// LowestCommonAncestorBST возвращает наименьшего общего предка узлов p и q
// в двоичном дереве поиска: самый глубокий узел, у которого в поддереве
// есть и p, и q. Узел считается предком самого себя, поэтому если p лежит
// в поддереве q, ответ q. Гарантируется: дерево непустое, значения
// различны, p и q есть в дереве (p может совпадать с q).
//
// Ожидаемая сложность: O(h). Пользуйся свойством BST: значения подскажут,
// в какое поддерево идти. Обход всего дерева на каждый запрос (O(n))
// не уложится в лимит на 200 000 запросов.
func LowestCommonAncestorBST(root, p, q *TreeNode) *TreeNode {
	cur := root
	for cur != nil {
		switch {
		case p.Val < cur.Val && q.Val < cur.Val:
			cur = cur.Left // оба левее
		case p.Val > cur.Val && q.Val > cur.Val:
			cur = cur.Right // оба правее
		default:
			return cur // разошлись или один из них это cur
		}
	}
	return nil
}
