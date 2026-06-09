import 'package:flutter_test/flutter_test.dart';
import 'package:blockchain_app/main.dart';

void main() {
  testWidgets('App smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const BlockchainApp());
    // Verify app renders without crashing
    expect(find.text('YuteBlockchain Dashboard'), findsOneWidget);
  });
}
