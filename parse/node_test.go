package parse

// func Test_matcher_match(t *testing.T) {
// 	type fields struct {
// 		m *matcher
// 	}
// 	type args struct {
// 		itter slice.Itter[Node]
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   parseMatch
// 	}{
// 		{
// 			"match variable decl",
// 			fields{
// 				m: newGroup(Node{NodeType: Assign}).
// 					match(Identifyer, SetOp, Identifyer, NewLine).
// 					or(Identifyer, SetOp, Value, NewLine),
// 			},
// 			args{
// 				itter: slice.NewIttr([]Node{
// 					{NodeType: Identifyer}, {NodeType: SetOp}, {NodeType: Value}, {NodeType: NewLine},
// 				}),
// 			},
// 			parseMatch{
// 				ok:    true,
// 				count: 4,
// 				nodes: []Node{{NodeType: Assign, Nodes: []Node{
// 					{NodeType: Identifyer}, {NodeType: SetOp}, {NodeType: Value}, {NodeType: NewLine},
// 				}}},
// 			},
// 		},
// 		{
// 			"match use block",
// 			fields{
// 				m: newRoot(UseKeyword).
// 					match(OpenBlock, NewLine, CloseBlock, NewLine).
// 					or(OpenBlock, CloseBlock, NewLine),
// 			},
// 			args{
// 				itter: slice.NewIttr([]Node{
// 					{NodeType: UseKeyword}, {NodeType: OpenBlock}, {NodeType: CloseBlock}, {NodeType: NewLine},
// 				}),
// 			},
// 			parseMatch{
// 				ok:    true,
// 				count: 4,
// 				nodes: []Node{{NodeType: UseKeyword, Nodes: []Node{
// 					{NodeType: OpenBlock}, {NodeType: CloseBlock}, {NodeType: NewLine},
// 				}}},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := tt.fields.m.parse(tt.args.itter); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("matcher.match() = \n%v, want \n%v", got, tt.want)
// 			}
// 		})
// 	}
// }
